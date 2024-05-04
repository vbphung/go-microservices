package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const webPort = "80"

type Config struct {
	Rabbit *amqp.Connection
}

func main() {
	rabbitConn, err := connect()
	if err != nil {
		log.Panicln("Cannot connect to RabbitMQ: ", err)
		os.Exit(1)
	}

	defer rabbitConn.Close()

	app := Config{
		Rabbit: rabbitConn,
	}

	log.Printf("Start broker service on port %s\n", webPort)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func connect() (*amqp.Connection, error) {
	cnts := int64(0)
	backOff := 1 * time.Second
	var conn *amqp.Connection

	for {
		c, err := amqp.Dial("amqp://guest:guest@rabbitmq")
		if err != nil {
			fmt.Println("RabbitMQ isn't ready yet...")
			cnts++
		} else {
			conn = c
			break
		}

		if cnts > 5 {
			fmt.Println("Timeout error...")
			return nil, err
		}

		log.Println("Backing off...")

		backOff = time.Duration(math.Pow(float64(cnts), 2)) * time.Second
		time.Sleep(backOff)
	}

	return conn, nil
}
