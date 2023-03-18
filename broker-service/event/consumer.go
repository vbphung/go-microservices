package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn      *amqp.Connection
	queueName string
}

type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func NewConsumer(conn *amqp.Connection) (Consumer, error) {
	csm := Consumer{
		conn: conn,
	}

	err := csm.setup()
	if err != nil {
		return Consumer{}, err
	}

	return csm, nil
}

func (csm *Consumer) Listen(topics []string) error {
	ch, err := csm.conn.Channel()
	if err != nil {
		return err
	}

	defer ch.Close()

	q, err := declareRandomQueue(ch)
	if err != nil {
		return err
	}

	for _, t := range topics {
		err = ch.QueueBind(q.Name, t, "logs_topic", false, nil)
		if err != nil {
			return err
		}
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	fr := make(chan bool)
	go func() {
		for m := range msgs {
			var pl Payload
			_ = json.Unmarshal(m.Body, &pl)

			go handlePayload(pl)
		}
	}()

	fmt.Printf("Waiting for msg [logs_topic, %s]", q.Name)

	<-fr

	return nil
}

func handlePayload(pl Payload) {
	switch pl.Name {
	case "auth":
	case "log", "event":
		err := logEvent(pl)
		if err != nil {
			log.Println(err)
		}
	default:
		err := logEvent(pl)
		if err != nil {
			log.Println(err)
		}
	}
}

func logEvent(pl Payload) error {
	data, err := json.MarshalIndent(pl, "", "\t")
	if err != nil {
		return err
	}

	logServiceUrl := "http://logger-service/log"

	req, err := http.NewRequest("POST", logServiceUrl, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	cl := &http.Client{}

	resp, err := cl.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("Error logging: %d", resp.StatusCode)
	}

	return nil
}

func (csm *Consumer) setup() error {
	ch, err := csm.conn.Channel()
	if err != nil {
		return err
	}

	return declareExchange(ch)
}
