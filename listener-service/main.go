package main

import (
	"fmt"
	"listener/event"
	"log"
	"math"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	// connect to rabbitmq

	conn, err := connect()
	if err != nil {
		log.Panicln("Cannot connect to RabbitMQ: ", err)
		os.Exit(1)
	}

	defer conn.Close()

	log.Println("Connected to RabbitMQ")

	// start listening to messages

	log.Println("Listen and consume RabbitMQ messages...")

	// create consumer

	csm, err := event.NewConsumer(conn)
	if err != nil {
		log.Panicf("Cannot create consumer: %s\n", err)
	}

	// watch the queue and consume events

	err = csm.Listen([]string{"log.INFO", "log.WARNING", "log.ERROR"})
	if err != nil {
		log.Panicf("Consumer cannot listen: %s\n", err)
	}
}

func connect() (*amqp.Connection, error) {
	cnts := int64(0)
	backOff := 1 * time.Second
	var conn *amqp.Connection

	for {
		c, err := amqp.Dial("amqp://guest:guest@localhost")
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
