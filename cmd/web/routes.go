package main

import "net/http"

func (app *application) routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/api/v1.0/version", app.version)
	mux.HandleFunc("/api/v1.0/stats", app.stats)

	return mux
}

