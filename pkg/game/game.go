package game

import (
	"errors"
	"math/rand"

	"github.com/akarasz/yahtzee/pkg/models"
)

const (
	// maxRoll shows how many rolls a player have in one of their turn.
	maxRoll int = 3

	// totalRounds is the number of turns for the game.
	totalRounds int = 13
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
	AddPlayer() error
	Roll() ([]*models.Dice, error)
	Toggle(diceIndex int) ([]*models.Dice, error)
	Score(c models.Category) error
	Snapshot() *models.Game
}

type Concrete struct {
	Player string
	Game   *models.Game
}

func New(player string, g *models.Game) Controller {
	return &Concrete{player, g}
}

func (c *Concrete) currentPlayer() *models.Player {
	return c.Game.Players[c.Game.CurrentPlayer]
}

// AddPlayer adds a new player with the given `name` and an empty score sheet to the game.
func (c *Concrete) AddPlayer() error {
	if c.Game.CurrentPlayer > 0 || c.Game.Round > 0 {
		return ErrAlreadyStarted
	}

	for _, p := range c.Game.Players {
		if p.Name == c.Player {
			return ErrPlayerAlreadyAdded
		}
	}

	c.Game.Players = append(c.Game.Players, &models.Player{c.Player, map[models.Category]int{}})

	return nil
}

// Roll rolls the dices and increment the roll counters.
func (c *Concrete) Roll() ([]*models.Dice, error) {
	if c.Player != c.currentPlayer().Name {
		return nil, ErrNotPlayersTurn
	}

	if c.Game.Round >= totalRounds {
		return nil, ErrGameOver
	}

	if c.Game.RollCount >= maxRoll {
		return nil, ErrOutOfRolls
	}

	for _, d := range c.Game.Dices {
		if d.Locked {
			continue
		}

		d.Value = rand.Intn(6) + 1
	}

	c.Game.RollCount++

	return c.Game.Dices, nil
}

// Score saves the points for the player in the given category and handles the counters.
func (c *Concrete) Score(category models.Category) error {
	if c.Player != c.currentPlayer().Name {
		return ErrNotPlayersTurn
	}

	if c.Game.Round >= totalRounds {
		return ErrGameOver
	}

	if c.Game.RollCount == 0 {
		return ErrNoRollYet
	}

	if _, ok := c.currentPlayer().ScoreSheet[category]; ok {
		return ErrCategoryAlreadyScored
	}

	s := 0
	switch category {
	case models.Ones:
		for _, d := range c.Game.Dices {
			if d.Value == 1 {
				s++
			}
		}
	case models.Twos:
		for _, d := range c.Game.Dices {
			if d.Value == 2 {
				s += 2
			}
		}
	case models.Threes:
		for _, d := range c.Game.Dices {
			if d.Value == 3 {
				s += 3
			}
		}
	case models.Fours:
		for _, d := range c.Game.Dices {
			if d.Value == 4 {
				s += 4
			}
		}
	case models.Fives:
		for _, d := range c.Game.Dices {
			if d.Value == 5 {
				s += 5
			}
		}
	case models.Sixes:
		for _, d := range c.Game.Dices {
			if d.Value == 6 {
				s += 6
			}
		}
	case models.ThreeOfAKind:
		occurrences := map[int]int{}
		for _, d := range c.Game.Dices {
			occurrences[d.Value]++
		}

		for k, v := range occurrences {
			if v >= 3 {
				s = 3 * k
			}
		}
	case models.FourOfAKind:
		occurrences := map[int]int{}
		for _, d := range c.Game.Dices {
			occurrences[d.Value]++
		}

		for k, v := range occurrences {
			if v >= 4 {
				s = 4 * k
			}
		}
	case models.FullHouse:
		one, oneCount, other := c.Game.Dices[0].Value, 1, 0
		for i := 1; i < len(c.Game.Dices); i++ {
			v := c.Game.Dices[i].Value

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
	case models.SmallStraight:
		hit := [6]bool{}
		for _, d := range c.Game.Dices {
			hit[d.Value-1] = true
		}

		if (hit[0] && hit[1] && hit[2] && hit[3]) ||
			(hit[1] && hit[2] && hit[3] && hit[4]) ||
			(hit[2] && hit[3] && hit[4] && hit[5]) {
			s = 30
		}
	case models.LargeStraight:
		hit := [6]bool{}
		for _, d := range c.Game.Dices {
			hit[d.Value-1] = true
		}

		if (hit[0] && hit[1] && hit[2] && hit[3] && hit[4]) ||
			(hit[1] && hit[2] && hit[3] && hit[4] && hit[5]) {
			s = 40
		}
	case models.Yahtzee:
		same := true
		for i := 0; i < len(c.Game.Dices)-1; i++ {
			same = same && c.Game.Dices[i].Value == c.Game.Dices[i+1].Value
		}

		if same {
			s = 50
		}
	case models.Chance:
		for _, d := range c.Game.Dices {
			s += d.Value
		}
	default:
		return ErrInvalidCategory
	}

	c.currentPlayer().ScoreSheet[category] = s

	if _, ok := c.currentPlayer().ScoreSheet[models.Bonus]; !ok {
		var total, types int
		for k, v := range c.currentPlayer().ScoreSheet {
			if k == models.Ones || k == models.Twos || k == models.Threes ||
				k == models.Fours || k == models.Fives || k == models.Sixes {
				types++
				total += v
			}
		}

		if types == 6 {
			if total >= 63 {
				c.currentPlayer().ScoreSheet[models.Bonus] = 35
			} else {
				c.currentPlayer().ScoreSheet[models.Bonus] = 0
			}
		}
	}

	for _, d := range c.Game.Dices {
		d.Locked = false
	}

	c.Game.RollCount = 0
	c.Game.CurrentPlayer = (c.Game.CurrentPlayer + 1) % len(c.Game.Players)
	if c.Game.CurrentPlayer == 0 {
		c.Game.Round++
	}

	return nil
}

// Snapshot returns the game oject.
func (c *Concrete) Snapshot() *models.Game {
	return c.Game
}

// Toggle locks and unlocks a dice so it will not get rolled.
func (c *Concrete) Toggle(diceIndex int) ([]*models.Dice, error) {
	if c.Player != c.currentPlayer().Name {
		return nil, ErrNotPlayersTurn
	}

	if diceIndex < 0 || 4 < diceIndex {
		return nil, ErrInvalidDice
	}

	if c.Game.Round >= totalRounds {
		return nil, ErrGameOver
	}

	if c.Game.RollCount == 0 {
		return nil, ErrNoRollYet
	}

	if c.Game.RollCount >= 3 {
		return nil, ErrOutOfRolls
	}

	c.Game.Dices[diceIndex].Locked = !c.Game.Dices[diceIndex].Locked

	return c.Game.Dices, nil
}
