package game

import (
	"errors"

	log "github.com/sirupsen/logrus"
)

// ErrAlreadyStarted error is returned when pre-game operation is
// requested on an already started game.
var ErrAlreadyStarted = errors.New("game already started")

// Box represents one field on the scoring sheet.
type Box struct {
	// Filled shows whether the Box has a value already or it's still
	// available to choose.
	Filled bool

	// Score keeps the value of the box if it is filled.
	Score int
}

// Sheet represents a standard scoring sheet.
type Sheet struct {
	Ones, Twos, Threes, Fours, Fives, Sixes Box
	ThreeOfAKind, FourOfAKind               Box
	FullHouse                               Box
	SmallStraight, LargeStraight            Box
	Yahtzee                                 Box
	Chance                                  Box
}

// Player contains all data representing a player.
type Player struct {
	Name   string
	Scores *Sheet
}

func newPlayer(name string) *Player {
	return &Player{name, &Sheet{}}
}

// Game contains all data representing a game.
type Game struct {
	Players []*Player

	// Round shows how many rounds were passed already.
	Round int

	// Current shows the index of the current player in the Players array.
	Current int
}

// AddPlayer adds a new player with the given `name` and an
// empty score sheet to the game.
func (g *Game) AddPlayer(name string) error {
	log.Debugf("adding a player with name %q", name)

	if g.Current > 0 || g.Round > 0 {
		return ErrAlreadyStarted
	}

	g.Players = append(g.Players, newPlayer(name))

	return nil
}

// New initializes an empty Game.
func New() *Game {
	return &Game{}
}
