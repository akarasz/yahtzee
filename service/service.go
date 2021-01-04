package service

import (
	"errors"
	"math/rand"

	"github.com/akarasz/yahtzee/models"
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

// Game contains the possible actions for a yahtzee game.
type Game interface {
	// AddPlayer add a user to the game.
	AddPlayer() (models.Game, error)

	// Roll gives a new value for the unlocked dices.
	Roll() (models.Game, error)

	// Lock enables or disables a dice to roll.
	Lock(dice int) (models.Game, error)

	// Score saves the points in the player's score sheet.
	Score(c models.Category) (models.Game, error)
}

// Provider returns a new Game service.
type Provider interface {
	// Create returns a Game service object
	Create(g models.Game, u models.User) Game
}

// Default is the implementation of yahtzee.
type Default struct {
	game models.Game
	user models.User
}

func NewProvider() *Default {
	return &Default{}
}

func (s *Default) Create(g models.Game, u models.User) Game {
	return &Default{
		game: g,
		user: u,
	}
}

func (s *Default) AddPlayer() (models.Game, error) {
	g := s.game

	if g.CurrentPlayer > 0 || g.Round > 0 {
		return g, ErrAlreadyStarted
	}

	for _, p := range g.Players {
		if p.User == s.user {
			return g, ErrPlayerAlreadyAdded
		}
	}

	g.Players = append(g.Players, models.NewPlayer(s.user))
	return g, nil
}

func (s *Default) Roll() (models.Game, error) {
	g := s.game

	if len(g.Players) == 0 {
		return g, ErrNotPlayersTurn
	}

	currentPlayer := g.Players[g.CurrentPlayer]

	if s.user != currentPlayer.User {
		return g, ErrNotPlayersTurn
	}

	if g.Round >= totalRounds {
		return g, ErrGameOver
	}

	if g.RollCount >= maxRoll {
		return g, ErrOutOfRolls
	}

	for _, d := range g.Dices {
		if d.Locked {
			continue
		}

		d.Value = rand.Intn(6) + 1
	}

	g.RollCount++

	return g, nil
}

func (s *Default) Lock(diceIndex int) (models.Game, error) {
	g := s.game

	if len(g.Players) == 0 {
		return g, ErrNotPlayersTurn
	}

	currentPlayer := g.Players[g.CurrentPlayer]

	if s.user != currentPlayer.User {
		return g, ErrNotPlayersTurn
	}

	if diceIndex < 0 || 4 < diceIndex {
		return g, ErrInvalidDice
	}

	if g.Round >= totalRounds {
		return g, ErrGameOver
	}

	if g.RollCount == 0 {
		return g, ErrNoRollYet
	}

	if g.RollCount >= 3 {
		return g, ErrOutOfRolls
	}

	g.Dices[diceIndex].Locked = !g.Dices[diceIndex].Locked

	return g, nil
}

func (s *Default) Score(c models.Category) (models.Game, error) {
	g := s.game

	if len(g.Players) == 0 {
		return g, ErrNotPlayersTurn
	}

	currentPlayer := g.Players[g.CurrentPlayer]

	if s.user != currentPlayer.User {
		return g, ErrNotPlayersTurn
	}

	if g.Round >= totalRounds {
		return g, ErrGameOver
	}

	if g.RollCount == 0 {
		return g, ErrNoRollYet
	}

	if _, ok := currentPlayer.ScoreSheet[c]; ok {
		return g, ErrCategoryAlreadyScored
	}

	dices := make([]int, 5)
	for i, d := range g.Dices {
		dices[i] = d.Value
	}

	score, err := Score(c, dices)
	if err != nil {
		return g, err
	}

	currentPlayer.ScoreSheet[c] = score

	if _, ok := currentPlayer.ScoreSheet[models.Bonus]; !ok {
		var total, types int
		for k, v := range currentPlayer.ScoreSheet {
			if k == models.Ones || k == models.Twos || k == models.Threes ||
				k == models.Fours || k == models.Fives || k == models.Sixes {
				types++
				total += v
			}
		}

		if total >= 63 {
			currentPlayer.ScoreSheet[models.Bonus] = 35
		} else if types == 6 {
			currentPlayer.ScoreSheet[models.Bonus] = 0
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

	return g, nil
}

// Score calculates the category value for the given dices.
func Score(category models.Category, dices []int) (int, error) {
	s := 0
	switch category {
	case models.Ones:
		for _, d := range dices {
			if d == 1 {
				s++
			}
		}
	case models.Twos:
		for _, d := range dices {
			if d == 2 {
				s += 2
			}
		}
	case models.Threes:
		for _, d := range dices {
			if d == 3 {
				s += 3
			}
		}
	case models.Fours:
		for _, d := range dices {
			if d == 4 {
				s += 4
			}
		}
	case models.Fives:
		for _, d := range dices {
			if d == 5 {
				s += 5
			}
		}
	case models.Sixes:
		for _, d := range dices {
			if d == 6 {
				s += 6
			}
		}
	case models.ThreeOfAKind:
		occurrences := map[int]int{}
		for _, d := range dices {
			occurrences[d]++
		}

		for k, v := range occurrences {
			if v >= 3 {
				s = 3 * k
			}
		}
	case models.FourOfAKind:
		occurrences := map[int]int{}
		for _, d := range dices {
			occurrences[d]++
		}

		for k, v := range occurrences {
			if v >= 4 {
				s = 4 * k
			}
		}
	case models.FullHouse:
		one, oneCount, other := dices[0], 1, 0
		for i := 1; i < len(dices); i++ {
			v := dices[i]

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
		for _, d := range dices {
			hit[d-1] = true
		}

		if (hit[0] && hit[1] && hit[2] && hit[3]) ||
			(hit[1] && hit[2] && hit[3] && hit[4]) ||
			(hit[2] && hit[3] && hit[4] && hit[5]) {
			s = 30
		}
	case models.LargeStraight:
		hit := [6]bool{}
		for _, d := range dices {
			hit[d-1] = true
		}

		if (hit[0] && hit[1] && hit[2] && hit[3] && hit[4]) ||
			(hit[1] && hit[2] && hit[3] && hit[4] && hit[5]) {
			s = 40
		}
	case models.Yahtzee:
		same := true
		for i := 0; i < len(dices)-1; i++ {
			same = same && dices[i] == dices[i+1]
		}

		if same {
			s = 50
		}
	case models.Chance:
		for _, d := range dices {
			s += d
		}
	default:
		return 0, ErrInvalidCategory
	}

	return s, nil
}
