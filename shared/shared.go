// Package shared holds the APOD fetching, caching, and HTTP handlers.
package shared

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/kelseyhightower/envconfig"
)

// cacheControl is set on API responses; APOD only changes once a day.
const cacheControl = "public, max-age=900, s-maxage=900"

type Config struct {
	Port               string        `envconfig:"PORT" default:"8080"`
	CORSAllowedOrigins string        `envconfig:"CORS_ALLOWED_ORIGINS" default:"*"`
	NASAAPIKey         string        `envconfig:"NASA_API_KEY" required:"true"`
	NASABaseURL        string        `envconfig:"NASA_BASE_URL" default:"https://api.nasa.gov/planetary/apod"`
	CacheTTL           time.Duration `envconfig:"CACHE_TTL" default:"15m"`
}

func NewConfig() (*Config, error) {
	var c Config
	if err := envconfig.Process("apod", &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// Response holds the fields NASA's APOD API can return. All are optional:
// video days omit hdurl, "other" days omit url, and copyright is often absent.
type Response struct {
	Copyright      string `json:"copyright,omitempty"`
	Date           string `json:"date,omitempty"`
	Explanation    string `json:"explanation,omitempty"`
	HdURL          string `json:"hdurl,omitempty"`
	MediaType      string `json:"media_type,omitempty"`
	ServiceVersion string `json:"service_version,omitempty"`
	Title          string `json:"title,omitempty"`
	URL            string `json:"url,omitempty"`
	ThumbnailURL   string `json:"thumbnail_url,omitempty"`
}

// image returns the best still-image URL to serve for /image, or "" when the
// day has no representable image (media_type "other", or a video with no thumb).
func (r *Response) image() string {
	if r.MediaType == "video" {
		return r.ThumbnailURL
	}
	if r.URL != "" {
		return r.URL
	}
	return r.HdURL
}

// Service fetches (and briefly caches) the current APOD, and proxies its image.
type Service struct {
	conf   *Config
	client *http.Client

	mu     sync.Mutex
	cache  *Response
	expiry time.Time
}

func New(conf *Config) *Service {
	return &Service{
		conf:   conf,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

var (
	defaultOnce sync.Once
	defaultSvc  *Service
	defaultErr  error
)

// Default returns a process-wide Service built from the environment, reused
// across warm serverless invocations.
func Default() (*Service, error) {
	defaultOnce.Do(func() {
		conf, err := NewConfig()
		if err != nil {
			defaultErr = err
			return
		}
		defaultSvc = New(conf)
	})
	return defaultSvc, defaultErr
}

// Routes builds the API router: CORS, gzip, and the /apod and /image endpoints.
func (s *Service) Routes() *chi.Mux {
	r := chi.NewRouter()
	r.Use(
		cors.Handler(s.corsOptions()),
		middleware.Compress(5),
	)
	r.Get("/apod", s.HandleApod())
	r.Get("/image", s.HandleImage())
	return r
}

// CORS wraps a handler with the configured cross-origin policy.
func (s *Service) CORS(next http.Handler) http.Handler {
	return cors.Handler(s.corsOptions())(next)
}

func (s *Service) corsOptions() cors.Options {
	origins := strings.Split(s.conf.CORSAllowedOrigins, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}
	return cors.Options{
		AllowedOrigins: origins,
		AllowedMethods: []string{http.MethodGet, http.MethodOptions},
		AllowedHeaders: []string{"Accept", "Content-Type"},
		MaxAge:         300,
	}
}

// Fetch returns the current APOD, served from a short-lived in-memory cache when
// warm. Sending no date lets NASA pick the latest published picture.
func (s *Service) Fetch(ctx context.Context) (*Response, error) {
	s.mu.Lock()
	if s.cache != nil && time.Now().Before(s.expiry) {
		cached := *s.cache
		s.mu.Unlock()
		return &cached, nil
	}
	s.mu.Unlock()

	endpoint, err := url.Parse(s.conf.NASABaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid NASA base url: %w", err)
	}
	q := endpoint.Query()
	q.Set("api_key", s.conf.NASAAPIKey)
	q.Set("thumbs", "true")
	endpoint.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "go-apod")

	resp, err := s.client.Do(req)
	if err != nil {
		// The raw error embeds the request URL, which contains the API key.
		log.Printf("apod: request failed: %s", s.redact(err))
		return nil, errors.New("could not reach the APOD API")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("APOD API returned status %d", resp.StatusCode)
	}

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.New("could not decode the APOD API response")
	}
	result.Copyright = strings.TrimSpace(result.Copyright)

	s.mu.Lock()
	s.cache = &result
	s.expiry = time.Now().Add(s.conf.CacheTTL)
	s.mu.Unlock()

	out := result
	return &out, nil
}

// HandleApod returns the current APOD metadata as JSON.
func (s *Service) HandleApod() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apod, err := s.Fetch(r.Context())
		if err != nil {
			log.Printf("apod: %s", s.redact(err))
			http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", cacheControl)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if err := json.NewEncoder(w).Encode(apod); err != nil {
			log.Printf("apod: encode: %v", err)
		}
	}
}

// HandleImage proxies the current image, or a video day's thumbnail (NASA's
// image host sends no CORS headers).
func (s *Service) HandleImage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apod, err := s.Fetch(r.Context())
		if err != nil {
			log.Printf("image: %s", s.redact(err))
			http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
			return
		}

		src := apod.image()
		if src == "" {
			http.Error(w, "no image available today", http.StatusNotFound)
			return
		}

		req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, src, nil)
		if err != nil {
			log.Printf("image: %v", err)
			http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
			return
		}
		resp, err := s.client.Do(req)
		if err != nil {
			log.Printf("image: %v", err)
			http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
			return
		}
		if ct := resp.Header.Get("Content-Type"); ct != "" {
			w.Header().Set("Content-Type", ct)
		}
		if resp.ContentLength >= 0 {
			w.Header().Set("Content-Length", strconv.FormatInt(resp.ContentLength, 10))
		}
		w.Header().Set("Cache-Control", cacheControl)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if _, err := io.Copy(w, resp.Body); err != nil {
			log.Printf("image: copy: %v", err)
		}
	}
}

// StaticHandler serves the embedded frontend with a sensible cache policy.
func StaticHandler(fsys http.FileSystem) http.HandlerFunc {
	fileServer := http.FileServer(fsys)
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=3600")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		fileServer.ServeHTTP(w, r)
	}
}

// redact strips the API key from an error string so it is safe to log.
func (s *Service) redact(err error) string {
	return strings.ReplaceAll(err.Error(), s.conf.NASAAPIKey, "REDACTED")
}
