package game

import (
	"errors"
	"math/rand"

	log "github.com/sirupsen/logrus"
)

const (
	// numberOfDices shows how many dices are used for a game.
	numberOfDices int = 5

	// maxRoll shows how many rolls a player have in one of their turn.
	maxRoll int = 3

	// totalRounds is the number of turns for the game.
	totalRounds int = 13
)

// Category represents the formations players try to roll.
type Category string

// Available categories
const (
	Ones   Category = "ones"
	Twos            = "twos"
	Threes          = "threes"
	Fours           = "fours"
	Fives           = "fives"
	Sixes           = "sixes"

	ThreeOfAKind  = "three-of-a-kind"
	FourOfAKind   = "four-of-a-kind"
	FullHouse     = "full-house"
	SmallStraight = "small-straight"
	LargeStraight = "large-straight"
	Yahtzee       = "yahtzee"
	Chance        = "chance"
)

var (
	// ErrAlreadyStarted returned when pre-game operation is requested on an already started
	// game.
	ErrAlreadyStarted = errors.New("game already started")

	// ErrNotPlayersTurn returned when the requested operator was not initiated by the player
	// who's turn it is.
	ErrNotPlayersTurn = errors.New("not the player's turn")

	// ErrOutOfRolls returned when the player cannot roll again.
	ErrOutOfRolls = errors.New("out of rolls")

	// ErrGameOver returned when the game is over.
	ErrGameOver = errors.New("game over")
)

// Dice represents a dice you use for the Game.
type Dice struct {
	value  int
	locked bool
}

func (d *Dice) roll() {
	d.value = rand.Intn(6) + 1
}

// Value returns the number on the face of the dice.
func (d *Dice) Value() int {
	return d.value
}

func newDice() *Dice {
	d := &Dice{
		value: 1,
	}
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

	// roll shows how many times the dices were rolled for the current user in this round.
	roll int
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

// Roll rolls the dices and increment the roll counters.
func (g *Game) Roll(p *Player) error {
	if p != g.players[g.current] {
		return ErrNotPlayersTurn
	}

	if g.round >= totalRounds {
		return ErrGameOver
	}

	if g.roll >= maxRoll {
		return ErrOutOfRolls
	}

	for _, d := range g.dices {
		if d.locked {
			continue
		}

		d.roll()
	}

	g.roll++

	return nil
}

// Score saves the points for the player in the given category and handles the counters.
func (g *Game) Score(p *Player, c Category) error {
	v := 0
	switch c {
	case Ones:
		for _, d := range g.dices {
			if d.value == 1 {
				v++
			}
		}
	case Twos:
		for _, d := range g.dices {
			if d.value == 2 {
				v += 2
			}
		}
	case Threes:
		for _, d := range g.dices {
			if d.value == 3 {
				v += 3
			}
		}
	case Fours:
		for _, d := range g.dices {
			if d.value == 4 {
				v += 4
			}
		}
	case Fives:
		for _, d := range g.dices {
			if d.value == 5 {
				v += 5
			}
		}
	case Sixes:
		for _, d := range g.dices {
			if d.value == 6 {
				v += 6
			}
		}
	}

	p.scoreSheet[c] = v

	return nil
}

// New initializes an empty Game.
func New() *Game {
	dd := make([]*Dice, numberOfDices)
	for i := 0; i < numberOfDices; i++ {
		dd[i] = newDice()
	}

	return &Game{
		dices: dd,
	}
}
