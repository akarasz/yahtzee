package game_test

import (
	"testing"

	"github.com/akarasz/yahtzee/pkg/game"
)

func TestNewGame(t *testing.T) {
	t.Run("should create with empty Players", func(t *testing.T) {
		got := game.New()

		if len(got.Players) > 0 {
			t.Errorf("NewGame() should produce empty Players list")
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
		if p.Scores == nil {
			t.Errorf("Scores is nil")
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
