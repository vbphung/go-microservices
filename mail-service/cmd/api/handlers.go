package main

import "net/http"

func (app *Config) SendMail(w http.ResponseWriter, r *http.Request) {
	type mailMsg struct {
		From    string `json:"from"`
		To      string `json:"to"`
		Subject string `json:"subject"`
		Message string `json:"message"`
	}

	var reqPl mailMsg
	err := app.readJson(w, r, &reqPl)
	if err != nil {
		app.errorJson(w, err)
		return
	}

	msg := Message{
		From:    reqPl.From,
		To:      reqPl.To,
		Subject: reqPl.Subject,
		Data:    reqPl.Message,
	}

	if err = app.Mailer.SendSMTPMessage(msg); err != nil {
		app.errorJson(w, err)
		return
	}

	app.writeJson(w, http.StatusAccepted, jsonResp{
		Error:   false,
		Message: "Sent to " + reqPl.To,
	})
}
