package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	"github.com/lissy93/go-apod/shared"
)

//go:embed static
var staticFiles embed.FS

func main() {
	conf, err := shared.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	content, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatal(err)
	}

	r := shared.New(conf).Routes()
	r.Get("/*", shared.StaticHandler(http.FS(content)))

	log.Printf("\033[1;92m🌌 Go-APOD running at http://localhost:%s/\033[0m", conf.Port)
	log.Fatal(http.ListenAndServe(":"+conf.Port, r))
}
