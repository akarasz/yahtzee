package events

//go:generate mockgen -destination=mocks/mock_events.go -package=mocks -build_flags=-mod=mod . Emitter

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

// Type tells which kind of events happened
type Type int

// Available types
const (
	AddPlayer Type = iota
	Roll
	Lock
	Score
)

// Emitter used by the event producer side to fire events
type Emitter interface {
	// Emit notifies the consumers of `gameID` that a `t` event happened
	// with the changes described in `body`
	Emit(gameID string, t Type, body interface{})
}

type LoggingEmitter struct{}

func (e *LoggingEmitter) Emit(gameID string, t Type, body interface{}) {
	jsonBody, _ := json.Marshal(body)
	log.WithFields(log.Fields{
		"gameID": gameID,
		"type":   t,
	}).Info(string(jsonBody))
}
