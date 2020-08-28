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
	AddPlayer(g *models.Game, name string) error
	Roll(g *models.Game, player string) ([]*models.Dice, error)
	Toggle(g *models.Game, player string, diceIndex int) ([]*models.Dice, error)
	Score(g *models.Game, player string, c models.Category) error
}

type implementation struct {
}

// New returns the default controller implementation.
func New() Controller {
	return &implementation{}
}

func (c *implementation) addPlayer(g *models.Game, name string) {
	g.Players = append(g.Players, &models.Player{
		Name:       name,
		ScoreSheet: map[models.Category]int{}},
	)
}

func (c *implementation) currentPlayer(g *models.Game) *models.Player {
	return g.Players[g.CurrentPlayer]
}

// AddPlayer adds a new player with the given `name` and an empty score sheet to the game.
func (c *implementation) AddPlayer(g *models.Game, name string) error {
	if g.CurrentPlayer > 0 || g.Round > 0 {
		return ErrAlreadyStarted
	}

	for _, p := range g.Players {
		if p.Name == name {
			return ErrPlayerAlreadyAdded
		}
	}

	c.addPlayer(g, name)

	return nil
}

// Roll rolls the dices and increment the roll counters.
func (c *implementation) Roll(g *models.Game, player string) ([]*models.Dice, error) {
	if len(g.Players) == 0 || player != c.currentPlayer(g).Name {
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

		d.Value = rand.Intn(6) + 1
	}

	g.RollCount++

	return g.Dices, nil
}

// Score saves the points for the player in the given category and handles the counters.
func (c *implementation) Score(g *models.Game, player string, category models.Category) error {
	if len(g.Players) == 0 || player != c.currentPlayer(g).Name {
		return ErrNotPlayersTurn
	}

	if g.Round >= totalRounds {
		return ErrGameOver
	}

	if g.RollCount == 0 {
		return ErrNoRollYet
	}

	if _, ok := c.currentPlayer(g).ScoreSheet[category]; ok {
		return ErrCategoryAlreadyScored
	}

	s := 0
	switch category {
	case models.Ones:
		for _, d := range g.Dices {
			if d.Value == 1 {
				s++
			}
		}
	case models.Twos:
		for _, d := range g.Dices {
			if d.Value == 2 {
				s += 2
			}
		}
	case models.Threes:
		for _, d := range g.Dices {
			if d.Value == 3 {
				s += 3
			}
		}
	case models.Fours:
		for _, d := range g.Dices {
			if d.Value == 4 {
				s += 4
			}
		}
	case models.Fives:
		for _, d := range g.Dices {
			if d.Value == 5 {
				s += 5
			}
		}
	case models.Sixes:
		for _, d := range g.Dices {
			if d.Value == 6 {
				s += 6
			}
		}
	case models.ThreeOfAKind:
		occurrences := map[int]int{}
		for _, d := range g.Dices {
			occurrences[d.Value]++
		}

		for k, v := range occurrences {
			if v >= 3 {
				s = 3 * k
			}
		}
	case models.FourOfAKind:
		occurrences := map[int]int{}
		for _, d := range g.Dices {
			occurrences[d.Value]++
		}

		for k, v := range occurrences {
			if v >= 4 {
				s = 4 * k
			}
		}
	case models.FullHouse:
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
	case models.SmallStraight:
		hit := [6]bool{}
		for _, d := range g.Dices {
			hit[d.Value-1] = true
		}

		if (hit[0] && hit[1] && hit[2] && hit[3]) ||
			(hit[1] && hit[2] && hit[3] && hit[4]) ||
			(hit[2] && hit[3] && hit[4] && hit[5]) {
			s = 30
		}
	case models.LargeStraight:
		hit := [6]bool{}
		for _, d := range g.Dices {
			hit[d.Value-1] = true
		}

		if (hit[0] && hit[1] && hit[2] && hit[3] && hit[4]) ||
			(hit[1] && hit[2] && hit[3] && hit[4] && hit[5]) {
			s = 40
		}
	case models.Yahtzee:
		same := true
		for i := 0; i < len(g.Dices)-1; i++ {
			same = same && g.Dices[i].Value == g.Dices[i+1].Value
		}

		if same {
			s = 50
		}
	case models.Chance:
		for _, d := range g.Dices {
			s += d.Value
		}
	default:
		return ErrInvalidCategory
	}

	c.currentPlayer(g).ScoreSheet[category] = s

	if _, ok := c.currentPlayer(g).ScoreSheet[models.Bonus]; !ok {
		var total, types int
		for k, v := range c.currentPlayer(g).ScoreSheet {
			if k == models.Ones || k == models.Twos || k == models.Threes ||
				k == models.Fours || k == models.Fives || k == models.Sixes {
				types++
				total += v
			}
		}

		if types == 6 {
			if total >= 63 {
				c.currentPlayer(g).ScoreSheet[models.Bonus] = 35
			} else {
				c.currentPlayer(g).ScoreSheet[models.Bonus] = 0
			}
		}
	}

	for _, d := range g.Dices {
		d.Locked = false
	}

	g.RollCount = 0
	g.CurrentPlayer = (g.CurrentPlayer + 1) % len(g.Players)
	if g.CurrentPlayer == 0 {
		g.Round++
	}

	return nil
}

// Toggle locks and unlocks a dice so it will not get rolled.
func (c *implementation) Toggle(g *models.Game, player string, diceIndex int) ([]*models.Dice, error) {
	if len(g.Players) == 0 || player != c.currentPlayer(g).Name {
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
