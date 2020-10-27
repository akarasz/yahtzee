package events

import (
	"github.com/streadway/amqp"
)

type Rabbit struct {
	conn *amqp.Connection
	ch *amqp.Channel
}

func (r *Rabbit) Close() {
	r.ch.Close()
	r.conn.Close()
}

func NewRabbit(uri string) (*Rabbit, error) {
	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &Rabbit{
		conn: conn,
		ch: ch,
	}, nil
}

func (r *Rabbit) Emit(gameID string, t Type, body interface{}) {
	amqp.Dial("a")
}
