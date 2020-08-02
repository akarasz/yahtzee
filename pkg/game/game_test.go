package game_test

import (
	"testing"

	"github.com/akarasz/yahtzee/pkg/game"
)

func TestNewGame(t *testing.T) {
	t.Run("should create with empty Players", func(t *testing.T) {
		g := game.New()

		if len(g.Players) != 0 {
			t.Errorf("NewGame() should produce empty Players list")
		}
	})

	t.Run("should add dices", func(t *testing.T) {
		g := game.New()

		if got, want := len(g.Dices), game.NumberOfDices; got != want {
			t.Errorf("number of dices is invalid, got %d, want %d.", got, want)
		}
	})

	t.Run("dices should have valid values", func(t *testing.T) {
		g := game.New()

		for i, d := range g.Dices {
			if got := d.Value(); got < 1 || got > 6 {
				t.Errorf("%dth dice has an invalid value, %d.", i, got)
			}
		}
	})
}

func TestGame(t *testing.T) {
	t.Run("addPlayer should add with empty sheet and give name", func(t *testing.T) {
		g := game.New()

		g.AddPlayer("alice")

		if len(g.Players) != 1 {
			t.Fatalf("player was not added")
		}
		p := g.Players[0]
		if p.Name != "alice" {
			t.Errorf("wrong Name %q", p.Name)
		}
		if len(p.ScoreSheet) != 0 {
			t.Errorf("ScoreSheet is not empty")
		}
	})

	t.Run("addPlayer should fail when game started", func(t *testing.T) {
		table := []struct {
			current, round int
			expected       error
		}{
			{0, 0, nil},
			{0, 1, game.ErrAlreadyStarted},
			{1, 0, game.ErrAlreadyStarted},
			{2, 3, game.ErrAlreadyStarted},
		}

		for _, row := range table {
			g := game.New()
			g.Current = row.current
			g.Round = row.round

			got := g.AddPlayer("alice")

			if got != row.expected {
				t.Errorf("adding to %v was incorrect, got: %v, want: %v.",
					g,
					got,
					row.expected)
			}
		}
	})
}
