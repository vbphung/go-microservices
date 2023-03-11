package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func (app *Config) Auth(w http.ResponseWriter, r *http.Request) {
	var reqPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJson(w, r, &reqPayload)
	if err != nil {
		app.errorJson(w, err, http.StatusBadRequest)
		return
	}

	user, err := app.Models.User.GetByEmail(reqPayload.Email)
	if err != nil {
		app.errorJson(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	valid, err := user.PasswordMatches(reqPayload.Password)
	if err != nil || !valid {
		app.errorJson(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	pl := jsonResp{
		Error:   false,
		Message: fmt.Sprintf("Logged with user %s", user.Email),
		Data:    user,
	}

	err = app.logReq("authentication", fmt.Sprintf("%s logged in successfully!", user.Email))
	if err != nil {
		app.errorJson(w, err)
		return
	}

	app.writeJson(w, http.StatusAccepted, pl)
}

func (app *Config) logReq(name, data string) error {
	logEntr := struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}{
		Name: name,
		Data: data,
	}

	jsonData, err := json.MarshalIndent(logEntr, "", "\t")
	if err != nil {
		return err
	}

	logServiceUrl := "http://log-service/log"

	req, err := http.NewRequest("POST", logServiceUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	client := &http.Client{}

	_, err = client.Do(req)
	if err != nil {
		return err
	}

	return nil
}
