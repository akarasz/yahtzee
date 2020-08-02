package game

import (
	"errors"
	"math/rand"

	log "github.com/sirupsen/logrus"
)

const (
	// NumberOfDices shows how many dices are used for a game.
	NumberOfDices int = 5
)

// Category represents the formations players try to roll.
type Category int

// Available categories
const (
	Ones Category = iota
	Twos
	Threes
	Fours
	Fives
	Sixes

	ThreeOfAKind
	FourOfAKind
	FullHouse
	SmallStraight
	LargeStraight
	Yahtzee
	Chance
)

// ErrAlreadyStarted error is returned when pre-game operation is requested on an already started
// game.
var ErrAlreadyStarted = errors.New("game already started")

// Dice represents a dice you use for the Game.
type Dice struct {
	value int
}

func (d *Dice) roll() {
	d.value = rand.Intn(6) + 1
}

// Value returns the number on the face of the dice.
func (d *Dice) Value() int {
	return d.value
}

func newDice() *Dice {
	d := &Dice{1}
	return d
}

// Player contains all data representing a player.
type Player struct {
	name       string
	scoreSheet map[Category]int
}

func newPlayer(name string) *Player {
	return &Player{name, map[Category]int{}}
}

// Game contains all data representing a game.
type Game struct {
	players []*Player

	dices []*Dice

	// round shows how many rounds were passed already.
	round int

	// current shows the index of the current player in the Players array.
	current int

	// reroll shows how many times the dices were rerolled for the current user in this round.
	reroll int
}

// AddPlayer adds a new player with the given `name` and an empty score sheet to the game.
func (g *Game) AddPlayer(name string) error {
	log.Debugf("adding a player with name %q", name)

	if g.current > 0 || g.round > 0 {
		return ErrAlreadyStarted
	}

	g.players = append(g.players, newPlayer(name))

	return nil
}

func (g *Game) Roll(p *Player) error {
	for _, d := range g.dices {
		d.roll()
	}

	return nil
}

func (g *Game) Score(p *Player, c Category) error {
	return nil
}

// New initializes an empty Game.
func New() *Game {
	dd := make([]*Dice, NumberOfDices)
	for i := 0; i < NumberOfDices; i++ {
		dd[i] = newDice()
	}

	return &Game{
		dices: dd,
	}
}
