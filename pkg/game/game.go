package game

import (
	"errors"
	"math/rand"
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
	Bonus           = "bonus"

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

	// ErrInvalidCategory returned when scoring category not valid.
	ErrInvalidCategory = errors.New("invalid category")

	// ErrCategoryAlreadyScored returned when the category in the player's score sheet is filled.
	ErrCategoryAlreadyScored = errors.New("category already scored")

	// ErrNoRollYet returned when there was no rolling yet.
	ErrNoRollYet = errors.New("dices should be rolled first")

	// ErrInvalidDice returned when dice index is invalid.
	ErrInvalidDice = errors.New("invalid dice")

	// ErrPlayerAlreadyAdded returned when the player is already added to the game.
	ErrPlayerAlreadyAdded = errors.New("player already added")
)

type Controller interface {
	AddPlayer(name string) error
	Roll(player string) ([]*Dice, error)
	Toggle(player string, diceIndex int) ([]*Dice, error)
	Score(player string, c Category) error
	Snapshot() *Game
}

// Dice represents a dice you use for the Game.
type Dice struct {
	// Value is the number on the face of the dice
	Value int

	// Locked shows if the dice will roll or not
	Locked bool
}

func (d *Dice) roll() {
	d.Value = rand.Intn(6) + 1
}

func newDice() *Dice {
	d := &Dice{
		Value: 1,
	}
	return d
}

// Player contains all data representing a player.
type Player struct {
	// Name of the player
	Name string

	// ScoreSheet keeps the scores of the player
	ScoreSheet map[Category]int
}

// Game contains all data representing a game.
type Game struct {
	// Players has the list of the players in an ordered manner
	Players []*Player

	// Dices has the dices the game played with
	Dices []*Dice

	// Round shows how many rounds were passed already.
	Round int

	// Current shows the index of the current player in the Players array.
	Current int

	// RollCount shows how many times the dices were rolled for the current user in this round.
	RollCount int
}

func (g *Game) currentPlayer() *Player {
	return g.Players[g.Current]
}

// AddPlayer adds a new player with the given `name` and an empty score sheet to the game.
func (g *Game) AddPlayer(name string) error {
	if g.Current > 0 || g.Round > 0 {
		return ErrAlreadyStarted
	}

	for _, p := range g.Players {
		if p.Name == name {
			return ErrPlayerAlreadyAdded
		}
	}

	g.Players = append(g.Players, &Player{name, map[Category]int{}})

	return nil
}

// Roll rolls the dices and increment the roll counters.
func (g *Game) Roll(player string) ([]*Dice, error) {
	if player != g.currentPlayer().Name {
		return nil, ErrNotPlayersTurn
	}

	if g.Round >= totalRounds {
		return nil, ErrGameOver
	}

	if g.RollCount >= maxRoll {
		return nil, ErrOutOfRolls
	}

	for _, d := range g.Dices {
		if d.Locked {
			continue
		}

		d.roll()
	}

	g.RollCount++

	return g.Dices, nil
}

// Score saves the points for the player in the given category and handles the counters.
func (g *Game) Score(player string, c Category) error {
	if player != g.currentPlayer().Name {
		return ErrNotPlayersTurn
	}

	if g.Round >= totalRounds {
		return ErrGameOver
	}

	if g.RollCount == 0 {
		return ErrNoRollYet
	}

	if _, ok := g.currentPlayer().ScoreSheet[c]; ok {
		return ErrCategoryAlreadyScored
	}

	s := 0
	switch c {
	case Ones:
		for _, d := range g.Dices {
			if d.Value == 1 {
				s++
			}
		}
	case Twos:
		for _, d := range g.Dices {
			if d.Value == 2 {
				s += 2
			}
		}
	case Threes:
		for _, d := range g.Dices {
			if d.Value == 3 {
				s += 3
			}
		}
	case Fours:
		for _, d := range g.Dices {
			if d.Value == 4 {
				s += 4
			}
		}
	case Fives:
		for _, d := range g.Dices {
			if d.Value == 5 {
				s += 5
			}
		}
	case Sixes:
		for _, d := range g.Dices {
			if d.Value == 6 {
				s += 6
			}
		}
	case ThreeOfAKind:
		occurrences := map[int]int{}
		for _, d := range g.Dices {
			occurrences[d.Value]++
		}

		for k, v := range occurrences {
			if v >= 3 {
				s = 3 * k
			}
		}
	case FourOfAKind:
		occurrences := map[int]int{}
		for _, d := range g.Dices {
			occurrences[d.Value]++
		}

		for k, v := range occurrences {
			if v >= 4 {
				s = 4 * k
			}
		}
	case FullHouse:
		one, oneCount, other := g.Dices[0].Value, 1, 0
		for i := 1; i < len(g.Dices); i++ {
			v := g.Dices[i].Value

			if one == v {
				oneCount++
			} else if other == 0 || other == v {
				other = v
			} else {
				oneCount = 4
			}
		}

		if oneCount == 2 || oneCount == 3 {
			s = 25
		}
	case SmallStraight:
		hit := [6]bool{}
		for _, d := range g.Dices {
			hit[d.Value-1] = true
		}

		if (hit[0] && hit[1] && hit[2] && hit[3]) ||
			(hit[1] && hit[2] && hit[3] && hit[4]) ||
			(hit[2] && hit[3] && hit[4] && hit[5]) {
			s = 30
		}
	case LargeStraight:
		hit := [6]bool{}
		for _, d := range g.Dices {
			hit[d.Value-1] = true
		}

		if (hit[0] && hit[1] && hit[2] && hit[3] && hit[4]) ||
			(hit[1] && hit[2] && hit[3] && hit[4] && hit[5]) {
			s = 40
		}
	case Yahtzee:
		same := true
		for i := 0; i < len(g.Dices)-1; i++ {
			same = same && g.Dices[i].Value == g.Dices[i+1].Value
		}

		if same {
			s = 50
		}
	case Chance:
		for _, d := range g.Dices {
			s += d.Value
		}
	default:
		return ErrInvalidCategory
	}

	g.currentPlayer().ScoreSheet[c] = s

	if _, ok := g.currentPlayer().ScoreSheet[Bonus]; !ok {
		var total, types int
		for k, v := range g.currentPlayer().ScoreSheet {
			if k == Ones || k == Twos || k == Threes || k == Fours || k == Fives || k == Sixes {
				types++
				total += v
			}
		}

		if types == 6 {
			if total >= 63 {
				g.currentPlayer().ScoreSheet[Bonus] = 35
			} else {
				g.currentPlayer().ScoreSheet[Bonus] = 0
			}
		}
	}

	for _, d := range g.Dices {
		d.Locked = false
	}

	g.RollCount = 0
	g.Current = (g.Current + 1) % len(g.Players)
	if g.Current == 0 {
		g.Round++
	}

	return nil
}

// Snapshot returns the game oject.
func (g *Game) Snapshot() *Game {
	return g
}

// Toggle locks and unlocks a dice so it will not get rolled.
func (g *Game) Toggle(player string, diceIndex int) ([]*Dice, error) {
	if player != g.currentPlayer().Name {
		return nil, ErrNotPlayersTurn
	}

	if diceIndex < 0 || 4 < diceIndex {
		return nil, ErrInvalidDice
	}

	if g.Round >= totalRounds {
		return nil, ErrGameOver
	}

	if g.RollCount == 0 {
		return nil, ErrNoRollYet
	}

	if g.RollCount >= 3 {
		return nil, ErrOutOfRolls
	}

	g.Dices[diceIndex].Locked = !g.Dices[diceIndex].Locked

	return g.Dices, nil
}

// New initializes an empty Game.
func New() *Game {
	dd := make([]*Dice, numberOfDices)
	for i := 0; i < numberOfDices; i++ {
		dd[i] = newDice()
	}

	return &Game{
		Dices: dd,
	}
}
