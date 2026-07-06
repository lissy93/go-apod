package shared

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const testKey = "super-secret-key"

// stubNASA stands in for the upstream API. It serves the APOD JSON (settable per
// test), plus /img and /thumb image endpoints, and counts JSON hits so tests can
// assert the cache is working.
func stubNASA(t *testing.T) (base string, hits *int32, setBody func(string)) {
	t.Helper()
	var count int32
	var body atomic.Value
	body.Store("{}")

	mux := http.NewServeMux()
	mux.HandleFunc("/apod", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&count, 1)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, body.Load().(string))
	})
	mux.HandleFunc("/img", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		io.WriteString(w, "IMAGE-BYTES")
	})
	mux.HandleFunc("/thumb", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		io.WriteString(w, "THUMB-BYTES")
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv.URL, &count, func(s string) { body.Store(s) }
}

func newService(base string) *Service {
	return New(&Config{
		NASAAPIKey:  testKey,
		NASABaseURL: base + "/apod",
		CacheTTL:    time.Minute,
	})
}

func TestFetchParsesAndTrimsCopyright(t *testing.T) {
	base, _, setBody := stubNASA(t)
	setBody(`{"media_type":"image","title":"Moon","url":"u","copyright":"  Jane Doe \n"}`)

	got, err := newService(base).Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if got.Title != "Moon" {
		t.Errorf("title = %q, want Moon", got.Title)
	}
	if got.Copyright != "Jane Doe" {
		t.Errorf("copyright = %q, want trimmed 'Jane Doe'", got.Copyright)
	}
}

func TestFetchCaches(t *testing.T) {
	base, hits, setBody := stubNASA(t)
	setBody(`{"media_type":"image","title":"Cached"}`)

	svc := newService(base)
	for i := 0; i < 3; i++ {
		if _, err := svc.Fetch(context.Background()); err != nil {
			t.Fatalf("Fetch %d: %v", i, err)
		}
	}
	if n := atomic.LoadInt32(hits); n != 1 {
		t.Errorf("upstream hit %d times, want 1 (cache should absorb the rest)", n)
	}
}

func TestHandleApod(t *testing.T) {
	base, _, setBody := stubNASA(t)
	setBody(`{"media_type":"image","title":"JSON Day","url":"u"}`)

	rr := httptest.NewRecorder()
	newService(base).HandleApod()(rr, httptest.NewRequest(http.MethodGet, "/apod", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("content-type = %q", ct)
	}
	if cc := rr.Header().Get("Cache-Control"); !strings.Contains(cc, "max-age") {
		t.Errorf("missing Cache-Control, got %q", cc)
	}
	var got Response
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Title != "JSON Day" {
		t.Errorf("title = %q", got.Title)
	}
}

func TestHandleImageServesImage(t *testing.T) {
	base, _, setBody := stubNASA(t)
	setBody(fmt.Sprintf(`{"media_type":"image","url":"%s/img"}`, base))

	rr := httptest.NewRecorder()
	newService(base).HandleImage()(rr, httptest.NewRequest(http.MethodGet, "/image", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if body := rr.Body.String(); body != "IMAGE-BYTES" {
		t.Errorf("body = %q, want proxied image", body)
	}
}

func TestHandleImageVideoServesThumbnail(t *testing.T) {
	base, _, setBody := stubNASA(t)
	setBody(fmt.Sprintf(`{"media_type":"video","url":"https://youtube/embed/x","thumbnail_url":"%s/thumb"}`, base))

	rr := httptest.NewRecorder()
	newService(base).HandleImage()(rr, httptest.NewRequest(http.MethodGet, "/image", nil))

	if body := rr.Body.String(); body != "THUMB-BYTES" {
		t.Errorf("video day should serve the thumbnail, got %q", body)
	}
}

func TestHandleImageOtherReturns404(t *testing.T) {
	base, _, setBody := stubNASA(t)
	setBody(`{"media_type":"other","title":"Interactive"}`)

	rr := httptest.NewRecorder()
	newService(base).HandleImage()(rr, httptest.NewRequest(http.MethodGet, "/image", nil))

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404 (no image on 'other' days)", rr.Code)
	}
}

// TestUpstreamErrorHidesAPIKey checks that a transport error, whose URL embeds
// the api_key, never reaches the client response or the returned error.
func TestUpstreamErrorHidesAPIKey(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	dead := l.Addr().String()
	l.Close() // nothing listens here now, so requests are refused

	svc := New(&Config{NASAAPIKey: testKey, NASABaseURL: "http://" + dead + "/apod", CacheTTL: time.Minute})

	if _, err := svc.Fetch(context.Background()); err == nil {
		t.Fatal("expected an error")
	} else if strings.Contains(err.Error(), testKey) {
		t.Errorf("returned error leaks the API key: %v", err)
	}

	rr := httptest.NewRecorder()
	svc.HandleApod()(rr, httptest.NewRequest(http.MethodGet, "/apod", nil))
	if rr.Code != http.StatusBadGateway {
		t.Errorf("status = %d, want 502", rr.Code)
	}
	if strings.Contains(rr.Body.String(), testKey) {
		t.Errorf("response body leaks the API key: %q", rr.Body.String())
	}
}

func TestRedact(t *testing.T) {
	svc := New(&Config{NASAAPIKey: testKey})
	err := fmt.Errorf(`Get "https://api.nasa.gov/planetary/apod?api_key=%s": timeout`, testKey)
	if got := svc.redact(err); strings.Contains(got, testKey) {
		t.Errorf("redact left the key in: %q", got)
	}
}
