package main

import "net/http"

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonResp{
		Error:   false,
		Message: "Hit the broker",
	}

	_ = app.writeJson(w, http.StatusOK, payload)
}
