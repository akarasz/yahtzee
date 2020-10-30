package events

import (
	"encoding/json"

	"github.com/streadway/amqp"
)

type Rabbit struct {
	conn *amqp.Connection
	ch   *amqp.Channel
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
		ch:   ch,
	}, nil
}

func (r *Rabbit) Emit(gameID string, t Type, body interface{}) {
	if err := r.exchangeDeclare(gameID); err != nil {
		return
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return
	}

	r.ch.Publish(
		gameID, // exchange
		"",     // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        jsonBody,
		})
}

func (r *Rabbit) Subscribe(gameID string, clientID interface{}) (chan interface{}, error) {
	if err := r.exchangeDeclare(gameID); err != nil {
		return nil, err
	}

	q, err := r.ch.QueueDeclare(
		"",    // name
		false, // durable
		true,  // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, err
	}

	err = r.ch.QueueBind(
		q.Name, // queue name
		"",     // routing key
		gameID, // exchange
		false,
		nil)
	if err != nil {
		return nil, err
	}

	// TODO create a chan and send all incoming messages to it until unsubscribe happens

	msgs, err := r.ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	c := make(chan interface{})
	go func() {
		for d := range msgs {
			c <- string(d.Body)
		}
	}()

	return c, nil
}

func (r *Rabbit) Unsubscribe(gameID string, clientID interface{}) error {
	return nil
}

func (r *Rabbit) exchangeDeclare(gameID string) error {
	return r.ch.ExchangeDeclare(
		gameID,   // name
		"fanout", // type
		true,     // durable
		true,     // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
}