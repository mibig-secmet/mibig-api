package main

import (
	"encoding/json"
	zap "go.uber.org/zap"
	"net/http"
)

func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *application) serverError(w http.ResponseWriter, err error) {
	app.logger.Errorw("server error", zap.Error(err))
	app.clientError(w, http.StatusInternalServerError)
}

func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

func (app *application) returnJson(v interface{}, w http.ResponseWriter) error {
	js, err := json.Marshal(v)
	if err != nil {
		app.logger.Errorw("error marshalling json",
			"error", err.Error(),
		)
		app.serverError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
	return nil
}
