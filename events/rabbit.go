package events

import (
	"github.com/streadway/amqp"
)

type RabbitMQ struct {
	conn *amqp.Connection
}

func (r *RabbitMQ) Emit(gameID string, t Type, body interface{}) {
	amqp.Dial()
}
