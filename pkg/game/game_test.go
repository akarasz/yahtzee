package game

import (
	"testing"
)

func TestNewGame(t *testing.T) {
	t.Run("should create with empty Players", func(t *testing.T) {
		g := New()

		if len(g.players) != 0 {
			t.Errorf("NewGame() should produce empty Players list")
		}
	})

	t.Run("should add dices", func(t *testing.T) {
		g := New()

		if got, want := len(g.dices), NumberOfDices; got != want {
			t.Errorf("number of dices is invalid, got %d, want %d.", got, want)
		}
	})

	t.Run("should set valid values for dices", func(t *testing.T) {
		g := New()

		for i, d := range g.dices {
			if got := d.Value(); got < 1 || got > 6 {
				t.Errorf("%dth dice has an invalid value, %d.", i, got)
			}
		}
	})
}

func TestGame_AddPlayer(t *testing.T) {
	t.Run("should add with empty sheet and give name", func(t *testing.T) {
		g := New()

		g.AddPlayer("alice")

		if len(g.players) != 1 {
			t.Fatalf("player was not added")
		}
		p := g.players[0]
		if p.name != "alice" {
			t.Errorf("wrong Name %q", p.name)
		}
		if len(p.scoreSheet) != 0 {
			t.Errorf("ScoreSheet is not empty")
		}
	})

	t.Run("should fail when game started", func(t *testing.T) {
		table := []struct {
			current, round int
			expected       error
		}{
			{0, 0, nil},
			{0, 1, ErrAlreadyStarted},
			{1, 0, ErrAlreadyStarted},
			{2, 3, ErrAlreadyStarted},
		}

		for _, row := range table {
			g := New()
			g.current = row.current
			g.round = row.round

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

func TestGame_Roll(t *testing.T) {
	t.Run("should set valid values for dices", func(t *testing.T) {
		g := New()
		for _, d := range g.dices {
			d.value = -1
		}

		g.Roll(&Player{})

		for i, d := range g.dices {
			if got := d.Value(); got < 1 || got > 6 {
				t.Errorf("%dth dice has an invalid value, %d.", i, got)
			}
		}
	})
}
