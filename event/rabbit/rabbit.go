package rabbit

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/streadway/amqp"

	"github.com/akarasz/yahtzee"
	"github.com/akarasz/yahtzee/event"
)

type Rabbit struct {
	ch *amqp.Channel

	sync.Mutex
	destroyChans map[interface{}]chan interface{}
}

func New(ch *amqp.Channel) (*Rabbit, error) {
	return &Rabbit{
		ch:           ch,
		destroyChans: map[interface{}]chan interface{}{},
	}, nil
}

func (r *Rabbit) Emit(gameID string, u *yahtzee.User, t event.Type, body interface{}) {
	if err := r.exchangeDeclare(gameID); err != nil {
		return
	}

	jsonBody, err := json.Marshal(event.Event{
		User:   u,
		Action: t,
		Data:   body,
	})
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

func (r *Rabbit) Subscribe(gameID string, clientID interface{}) (chan *event.Event, error) {
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

	c := make(chan *event.Event)
	d := make(chan interface{})
	r.Lock()
	r.destroyChans[clientID] = d
	r.Unlock()
	go func() {
		for {
			select {
			case m := <-msgs:
				var e event.Event
				if err := json.Unmarshal(m.Body, &e); err != nil {
					log.Printf("unable to unmarshal event: %v: %q", err, string(m.Body))
				} else {
					c <- &e
				}
			case <-d:
				return
			}
		}
	}()

	return c, nil
}

func (r *Rabbit) Unsubscribe(gameID string, clientID interface{}) error {
	r.Lock()
	if d, ok := r.destroyChans[clientID]; ok {
		d <- struct{}{}
		delete(r.destroyChans, clientID)
	}
	r.Unlock()

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
