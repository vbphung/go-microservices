package event

import (
	"context"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Emitter struct {
	conn *amqp.Connection
}

func (e *Emitter) setup() error {
	ch, err := e.conn.Channel()
	if err != nil {
		return err
	}

	defer ch.Close()

	return declareExchange(ch)
}

func (e *Emitter) Push(ev string, severity string) error {
	ch, err := e.conn.Channel()
	if err != nil {
		return err
	}

	defer ch.Close()

	log.Println("Pushing to channel...")

	if err = ch.PublishWithContext(context.Background(), "logs_topic", severity, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(ev),
	}); err != nil {
		return err
	}

	return nil
}

func NewEventEmitter(conn *amqp.Connection) (Emitter, error) {
	emt := Emitter{
		conn: conn,
	}

	err := emt.setup()
	if err != nil {
		return Emitter{}, err
	}

	return emt, nil
}
