package models

import (
	"testing"
)

func TestNewGame(t *testing.T) {
	t.Run("should create with empty Players", func(t *testing.T) {
		g := NewGame()

		if len(g.Players) != 0 {
			t.Errorf("NewGame() should produce empty Players list")
		}
	})

	t.Run("should add dices", func(t *testing.T) {
		g := NewGame()

		if got, want := len(g.Dices), 5; got != want {
			t.Errorf("number of dices is invalid, got %d, want %d.", got, want)
		}
	})

	t.Run("should set valid values for dices", func(t *testing.T) {
		g := NewGame()

		for i, d := range g.Dices {
			if got := d.Value; got < 1 || got > 6 {
				t.Errorf("%dth dice has an invalid value, %d.", i, got)
			}
		}
	})
}
