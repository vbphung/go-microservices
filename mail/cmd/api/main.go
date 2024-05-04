package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Config struct {
	Mailer Mail
}

const webPort = "80"

func main() {
	m, err := createMailer()
	if err != nil {
		log.Panicln("Cannot create mailer: ", err)
	}

	app := Config{
		Mailer: m,
	}

	log.Println("Start mail service: ", webPort)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Panicln("Cannot start server: ", err)
	}
}

func createMailer() (Mail, error) {
	port, err := strconv.Atoi(os.Getenv("MAIL_PORT"))
	if err != nil {
		return Mail{}, err
	}

	m := Mail{
		Domain:      os.Getenv("MAIL_DOMAIN"),
		Host:        os.Getenv("MAIL_HOST"),
		Port:        port,
		Username:    os.Getenv("MAIL_USERNAME"),
		Password:    os.Getenv("MAIL_PASSWORD"),
		Encryption:  os.Getenv("MAIL_ENCRYPTION"),
		FromName:    os.Getenv("MAIL_FROM_NAME"),
		FromAddress: os.Getenv("MAIL_FROM_ADDRESS"),
	}

	return m, nil
}
