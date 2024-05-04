package event

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

func declareExchange(ch *amqp.Channel) error {
	return ch.ExchangeDeclare(
		"logs_topic", // name
		"topic",      // type
		true,         // durable?
		false,        // autoDeleted?
		false,        // internal?
		false,        // noWait?
		nil,          // args?
	)
}

func declareRandomQueue(ch *amqp.Channel) (amqp.Queue, error) {
	return ch.QueueDeclare(
		"random_queue", // name
		false,          // durable?
		false,          // deleteWhenUnused?
		true,           // exclusive?
		false,          // noWait?
		nil,            // args?
	)
}
