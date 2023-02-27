package main

import (
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

	app.writeJson(w, http.StatusAccepted, pl)
}
