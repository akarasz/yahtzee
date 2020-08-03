package game

import (
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("should create with empty Players", func(t *testing.T) {
		g := New()

		if len(g.players) != 0 {
			t.Errorf("NewGame() should produce empty Players list")
		}
	})

	t.Run("should add dices", func(t *testing.T) {
		g := New()

		if got, want := len(g.dices), numberOfDices; got != want {
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

			if want := row.expected; got != want {
				t.Errorf("adding to %v was incorrect, got: %#v, want: %#v.", g, got, want)
			}
		}
	})
}

func TestGame_Roll(t *testing.T) {
	t.Run("should set valid values for dices", func(t *testing.T) {
		g := New()
		g.AddPlayer("alice")
		for _, d := range g.dices {
			d.value = -1
		}

		got := g.Roll(g.players[0])

		if got != nil {
			t.Errorf("returned an error: [%v]", got)
		}
		for i, d := range g.dices {
			if got := d.Value(); got < 1 || got > 6 {
				t.Errorf("%dth dice has an invalid value, %d.", i, got)
			}
		}
	})

	t.Run("should not roll locked dices", func(t *testing.T) {
		g := New()
		g.AddPlayer("alice")
		g.dices[2].locked = true
		g.dices[2].value = -1

		got := g.Roll(g.players[0])

		if got != nil {
			t.Errorf("returned an error: [%v]", got)
		}
		if g.dices[2].value != -1 {
			t.Errorf("value of locked dice got changed")
		}
	})

	t.Run("should increment roll count", func(t *testing.T) {
		g := New()
		g.AddPlayer("alice")

		got := g.Roll(g.players[0])

		if got != nil {
			t.Errorf("returned an error: [%v]", got)
		}
		if want := 1; g.roll != want {
			t.Errorf("not got incremented")
		}
	})

	t.Run("should return error when not player's turn", func(t *testing.T) {
		g := New()
		g.AddPlayer("alice")
		g.AddPlayer("bob")
		g.current = 1

		got := g.Roll(g.players[0])

		if want := ErrNotPlayersTurn; got != want {
			t.Errorf("wrong result, got %#v wanted %#v.", got, want)
		}
	})

	t.Run("should return error when out of rolls", func(t *testing.T) {
		g := New()
		g.AddPlayer("alice")
		g.roll = 3

		got := g.Roll(g.players[0])

		if want := ErrOutOfRolls; got != want {
			t.Errorf("wrong result, got %#v wanted %#v.", got, want)
		}
	})

	t.Run("should return error when out of rounds", func(t *testing.T) {
		g := New()
		g.AddPlayer("alice")
		g.round = 13

		got := g.Roll(g.players[0])

		if want := ErrGameOver; got != want {
			t.Errorf("wrong result, got %#v wanted %#v.", got, want)
		}
	})
}

func TestGame_Scroll(t *testing.T) {
	t.Run("should calculate points correctly", func(t *testing.T) {
		table := []struct {
			dices    []int
			category Category
			value    int
		}{
			{[]int{1, 2, 3, 1, 1}, Ones, 3},
			{[]int{2, 3, 4, 2, 3}, Twos, 4},
			{[]int{6, 4, 2, 2, 3}, Threes, 3},
			{[]int{1, 6, 3, 3, 5}, Fours, 0},
			{[]int{4, 4, 1, 2, 4}, Fours, 12},
			{[]int{6, 6, 3, 5, 2}, Fives, 5},
			{[]int{5, 3, 6, 6, 6}, Sixes, 18},
			{[]int{2, 4, 3, 6, 4}, ThreeOfAKind, 0},
			{[]int{3, 1, 3, 1, 3}, ThreeOfAKind, 9},
			{[]int{5, 2, 5, 5, 5}, ThreeOfAKind, 15},
			{[]int{2, 6, 3, 2, 2}, FourOfAKind, 0},
			{[]int{1, 6, 6, 6, 6}, FourOfAKind, 24},
			{[]int{4, 4, 4, 4, 4}, FourOfAKind, 16},
			{[]int{5, 5, 2, 5, 5}, FullHouse, 0},
			{[]int{2, 5, 3, 6, 5}, FullHouse, 0},
			{[]int{5, 5, 2, 5, 2}, FullHouse, 25},
			{[]int{3, 1, 3, 1, 3}, FullHouse, 25},
			{[]int{6, 2, 5, 1, 3}, SmallStraight, 0},
			{[]int{6, 2, 4, 1, 3}, SmallStraight, 30},
			{[]int{4, 2, 3, 5, 3}, SmallStraight, 30},
			{[]int{1, 6, 3, 5, 4}, SmallStraight, 30},
			{[]int{3, 5, 2, 3, 4}, LargeStraight, 0},
			{[]int{3, 5, 2, 1, 4}, LargeStraight, 40},
			{[]int{5, 2, 6, 3, 4}, LargeStraight, 40},
			{[]int{3, 3, 3, 3, 3}, Yahtzee, 50},
			{[]int{1, 1, 1, 1, 1}, Yahtzee, 50},
			{[]int{6, 2, 4, 1, 3}, Chance, 16},
			{[]int{1, 6, 3, 3, 5}, Chance, 18},
			{[]int{2, 3, 4, 2, 3}, Chance, 14},
		}

		for _, row := range table {
			g := New()
			g.AddPlayer("alice")
			for i, v := range row.dices {
				g.dices[i].value = v
			}

			got := g.Score(g.players[0], row.category)

			if got != nil {
				t.Errorf("returned an error: [%v]", got)
			}
			if got, want := g.players[0].scoreSheet[row.category], row.value; got != want {
				t.Errorf("%q score for [%v] should be %d but was %d.",
					row.category,
					row.dices,
					want,
					got)
			}
		}
	})

	t.Run("should set bonus if upper section reaches limit", func(t *testing.T) {
		table := []struct {
			values    []int
			remaining Category
			dices     []int
			bonus     bool
		}{
			{[]int{3, 6, -1, 16, 25, -1}, Sixes, []int{1, 3, 6, 2, 4}, false},
			{[]int{-1, -1, 12, -1, 20, 36}, Fours, []int{1, 3, 6, 2, 4}, false},
			{[]int{3, 6, 9, 16, 25, -1}, Sixes, []int{1, 3, 6, 2, 4}, true},
			{[]int{-1, 2, 3, 4, 15, 36}, Ones, []int{1, 1, 3, 3, 3}, false},
			{[]int{-1, 2, 3, 4, 15, 36}, Ones, []int{1, 1, 1, 3, 3}, true},
			{[]int{-1, 2, 3, 4, 15, 36}, Ones, []int{1, 1, 1, 1, 3}, true},
		}

		for i, row := range table {
			g := New()
			g.AddPlayer("alice")
			s := g.players[0].scoreSheet
			if row.values[0] > 0 {
				s[Ones] = row.values[0]
			}
			if row.values[1] > 0 {
				s[Twos] = row.values[1]
			}
			if row.values[2] > 0 {
				s[Threes] = row.values[2]
			}
			if row.values[3] > 0 {
				s[Fours] = row.values[3]
			}
			if row.values[4] > 0 {
				s[Fives] = row.values[4]
			}
			if row.values[5] > 0 {
				s[Sixes] = row.values[5]
			}
			for j, d := range g.dices {
				d.value = row.dices[j]
			}

			got := g.Score(g.players[0], row.remaining)

			if got != nil {
				t.Errorf("returned an error: [%v]", got)
			}
			if got, want := s[Bonus] == 35, row.bonus; got != want {
				t.Errorf("invalid result for scenario %d", i)
			}
		}
	})

	t.Run("should reset roll counter", func(t *testing.T) {
		// TODO
	})

	t.Run("should switch current to next player", func(t *testing.T) {
		// TODO
	})

	t.Run("should increment round when first player comes again", func(t *testing.T) {
		// TODO
	})

	t.Run("should fail when got invalid category", func(t *testing.T) {
		g := New()
		g.AddPlayer("alice")

		got := g.Score(g.players[0], Category("fake"))

		if want := ErrInvalidCategory; got != want {
			t.Errorf("wrong result, got %#v wanted %#v.", got, want)
		}
	})

	t.Run("should fail when category was already scored", func(t *testing.T) {
		g := New()
		g.AddPlayer("alice")
		g.players[0].scoreSheet[Twos] = 4

		got := g.Score(g.players[0], Twos)

		if want := ErrCategoryAlreadyScored; got != want {
			t.Errorf("wrong result, got %#v wanted %#v.", got, want)
		}
	})

	t.Run("should fail when game is over", func(t *testing.T) {
		// TODO
	})

	t.Run("should fail when there was no roll", func(t *testing.T) {
		// TODO
	})
}
