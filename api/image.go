package handler

import (
	"net/http"

	"github.com/lissy93/go-apod/shared"
)

// HandlerImage is the Vercel serverless entrypoint for /image.
func HandlerImage(w http.ResponseWriter, r *http.Request) {
	svc, err := shared.Default()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	svc.CORS(svc.HandleImage()).ServeHTTP(w, r)
}
