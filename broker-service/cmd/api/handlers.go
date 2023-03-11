package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
	Log    LogPayload  `json:"log,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonResp{
		Error:   false,
		Message: "Hit the broker",
	}

	_ = app.writeJson(w, http.StatusOK, payload)
}

func (app Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var reqPl RequestPayload

	err := app.readJson(w, r, &reqPl)
	if err != nil {
		app.errorJson(w, err)
		return
	}

	switch reqPl.Action {
	case "auth":
		app.authenticate(w, reqPl.Auth)
	case "log":
		app.log(w, reqPl.Log)
	default:
		app.errorJson(w, errors.New("unknown action"))
	}
}

func (app *Config) log(w http.ResponseWriter, logPl LogPayload) {
	data, err := json.MarshalIndent(logPl, "", "\t")
	if err != nil {
		app.errorJson(w, err)
		return
	}

	logServiceUrl := "http://log-service/log"

	req, err := http.NewRequest("POST", logServiceUrl, bytes.NewBuffer(data))
	if err != nil {
		app.errorJson(w, err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		app.errorJson(w, err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		app.errorJson(w, err)
		return
	}

	app.writeJson(w, http.StatusAccepted, jsonResp{
		Error:   false,
		Message: "log successfully",
	})
}

func (app *Config) authenticate(w http.ResponseWriter, auth AuthPayload) {
	data, err := json.MarshalIndent(auth, "", "\t")
	if err != nil {
		app.errorJson(w, err)
	}

	req, err := http.NewRequest("POST", "http://authentication-service/auth", bytes.NewBuffer(data))
	if err != nil {
		app.errorJson(w, err)
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		app.errorJson(w, err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		app.errorJson(w, errors.New("invalid credentials"))
		return
	} else if resp.StatusCode != http.StatusAccepted {
		app.errorJson(w, errors.New("error call auth service"))
		return
	}

	var respJson jsonResp

	err = json.NewDecoder(resp.Body).Decode(&respJson)
	if err != nil {
		app.errorJson(w, err)
		return
	}

	if respJson.Error {
		app.errorJson(w, errors.New(respJson.Message), http.StatusUnauthorized)
		return
	}

	pl := jsonResp{
		Error:   false,
		Message: "Authenticated",
		Data:    respJson.Data,
	}

	app.writeJson(w, http.StatusAccepted, pl)
}
