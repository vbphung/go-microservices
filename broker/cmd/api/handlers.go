package main

import (
	"broker/event"
	"broker/logs"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/rpc"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
	Log    LogPayload  `json:"log,omitempty"`
	Mail   MailPayload `json:"mail,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
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
		// app.rabbitLog(w, reqPl.Log)
		app.rpcLog(w, reqPl.Log)
	case "mail":
		app.sendMail(w, reqPl.Mail)
	default:
		app.errorJson(w, errors.New("unknown action"))
	}
}

func (app *Config) sendMail(w http.ResponseWriter, mailPl MailPayload) {
	data, err := json.MarshalIndent(mailPl, "", "\t")
	if err != nil {
		app.errorJson(w, err)
		return
	}

	mailServiceUrl := "http://mail-service/send"

	req, err := http.NewRequest("POST", mailServiceUrl, bytes.NewBuffer(data))
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
		Message: "Sent mail to " + mailPl.To,
	})
}

type RpcPayload struct {
	Name string
	Data string
}

func (app *Config) gRpcLog(w http.ResponseWriter, req *http.Request) {
	var reqPl RequestPayload

	err := app.readJson(w, req, &reqPl)
	if err != nil {
		app.errorJson(w, err)
		return
	}

	conn, err := grpc.Dial("logger-service:50001", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.errorJson(w, err)
		return
	}

	defer conn.Close()

	cl := logs.NewLogServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	_, err = cl.WriteLog(ctx, &logs.LogReq{
		LogEntry: &logs.Log{
			Name: reqPl.Log.Name,
			Data: reqPl.Log.Data,
		},
	})
	if err != nil {
		app.errorJson(w, err)
		return
	}

	app.writeJson(w, http.StatusAccepted, jsonResp{
		Error:   false,
		Message: "Logged via gRpc!",
	})
}

func (app *Config) rpcLog(w http.ResponseWriter, logPl LogPayload) {
	cl, err := rpc.Dial("tcp", "logger-service:5001")
	if err != nil {
		app.errorJson(w, err)
		return
	}

	var resp string

	if err = cl.Call("RpcServer.Log", RpcPayload(logPl), &resp); err != nil {
		app.errorJson(w, err)
		return
	}

	app.writeJson(w, http.StatusAccepted, jsonResp{
		Error:   false,
		Message: "Logged via Rpc!",
	})
}

func (app *Config) log(w http.ResponseWriter, logPl LogPayload) {
	data, err := json.MarshalIndent(logPl, "", "\t")
	if err != nil {
		app.errorJson(w, err)
		return
	}

	logServiceUrl := "http://logger-service/log"

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
		Message: "Log successfully",
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

func (app *Config) rabbitLog(w http.ResponseWriter, l LogPayload) {
	err := app.pushToQueue(l.Name, l.Data)
	if err != nil {
		app.errorJson(w, err)
		return
	}

	app.writeJson(w, http.StatusAccepted, jsonResp{
		Error:   false,
		Message: "Logged via Rabbit",
	})
}

func (app *Config) pushToQueue(name, msg string) error {
	emt, err := event.NewEventEmitter(app.Rabbit)
	if err != nil {
		return err
	}

	pl := LogPayload{
		Name: name,
		Data: msg,
	}

	j, err := json.MarshalIndent(&pl, "", "\t")
	if err != nil {
		return err
	}

	err = emt.Push(string(j), "log.INFO")
	if err != nil {
		return err
	}

	return nil
}
