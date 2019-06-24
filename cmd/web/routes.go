package main

import "net/http"

func (app *application) routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/api/v1/version", app.version)
	mux.HandleFunc("/api/v1/stats", app.stats)
	mux.HandleFunc("/api/v1/repository", app.repository)

	return mux
}
