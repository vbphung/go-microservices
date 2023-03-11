package main

import (
	"logger/data"
	"net/http"
)

type JsonPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) WriteLog(w http.ResponseWriter, r *http.Request) {
	var reqPl JsonPayload
	_ = app.readJson(w, r, &reqPl)

	ev := data.LogEntry{
		Name: reqPl.Name,
		Data: reqPl.Data,
	}

	err := data.Insert(ev)
	if err != nil {
		app.errorJson(w, err)
		return
	}

	app.writeJson(w, http.StatusAccepted, jsonResp{
		Error:   false,
		Message: "log successfully",
	})
}
