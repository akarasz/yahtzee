package game

import (
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("should create with empty Players", func(t *testing.T) {
		g := New()

		if len(g.Players) != 0 {
			t.Errorf("NewGame() should produce empty Players list")
		}
	})

	t.Run("should add dices", func(t *testing.T) {
		g := New()

		if got, want := len(g.Dices), 5; got != want {
			t.Errorf("number of dices is invalid, got %d, want %d.", got, want)
		}
	})

	t.Run("should set valid values for dices", func(t *testing.T) {
		g := New()

		for i, d := range g.Dices {
			if got := d.Value; got < 1 || got > 6 {
				t.Errorf("%dth dice has an invalid value, %d.", i, got)
			}
		}
	})
}

func TestGame_AddPlayer(t *testing.T) {
	t.Run("should add player", func(t *testing.T) {
		g := New()

		g.AddPlayer("alice")

		if len(g.Players) != 1 {
			t.Fatalf("player was not added")
		}
		if got, want := g.Players[0].Name, "alice"; got != want {
			t.Errorf("got [%v], want [%v]", got, want)
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
			g.Current = row.current
			g.Round = row.round

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
		for _, d := range g.Dices {
			d.Value = -1
		}

		got := g.Roll("alice")

		if got != nil {
			t.Fatalf("returned error: [%v]", got)
		}
		for i, d := range g.Dices {
			if got := d.Value; got < 1 || got > 6 {
				t.Errorf("%dth dice has an invalid value, %d.", i, got)
			}
		}
	})

	t.Run("should not roll locked dices", func(t *testing.T) {
		g := New()
		g.AddPlayer("alice")
		g.Dices[2].Locked = true
		g.Dices[2].Value = -1

		got := g.Roll("alice")

		if got != nil {
			t.Fatalf("returned error: [%v]", got)
		}
		if g.Dices[2].Value != -1 {
			t.Errorf("value of locked dice got changed")
		}
	})

	t.Run("should increment roll count", func(t *testing.T) {
		g := New()
		g.AddPlayer("alice")

		got := g.Roll("alice")

		if got != nil {
			t.Fatalf("returned error: [%v]", got)
		}
		if want := 1; g.RollCount != want {
			t.Errorf("not got incremented")
		}
	})

	t.Run("should return error when not player's turn", func(t *testing.T) {
		g := New()
		g.AddPlayer("alice")
		g.AddPlayer("bob")
		g.Current = 1

		got := g.Roll("alice")

		if want := ErrNotPlayersTurn; got != want {
			t.Errorf("wrong result, got %#v wanted %#v.", got, want)
		}
	})

	t.Run("should return error when out of rolls", func(t *testing.T) {
		g := New()
		g.AddPlayer("alice")
		g.RollCount = 3

		got := g.Roll("alice")

		if want := ErrOutOfRolls; got != want {
			t.Errorf("wrong result, got %#v wanted %#v.", got, want)
		}
	})

	t.Run("should return error when out of rounds", func(t *testing.T) {
		g := New()
		g.AddPlayer("alice")
		g.Round = 13

		got := g.Roll("alice")

		if want := ErrGameOver; got != want {
			t.Errorf("wrong result, got %#v wanted %#v.", got, want)
		}
	})
}

func TestGame_Score(t *testing.T) {
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
			g.Roll("alice")
			for i, v := range row.dices {
				g.Dices[i].Value = v
			}

			got := g.Score("alice", row.category)

			if got != nil {
				t.Fatalf("returned error: [%v]", got)
			}
			if got, want := g.Players[0].ScoreSheet[row.category], row.value; got != want {
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
			g.Roll("alice")
			alice := g.Players[0]
			if row.values[0] > 0 {
				alice.ScoreSheet[Ones] = row.values[0]
			}
			if row.values[1] > 0 {
				alice.ScoreSheet[Twos] = row.values[1]
			}
			if row.values[2] > 0 {
				alice.ScoreSheet[Threes] = row.values[2]
			}
			if row.values[3] > 0 {
				alice.ScoreSheet[Fours] = row.values[3]
			}
			if row.values[4] > 0 {
				alice.ScoreSheet[Fives] = row.values[4]
			}
			if row.values[5] > 0 {
				alice.ScoreSheet[Sixes] = row.values[5]
			}
			for j, d := range g.Dices {
				d.Value = row.dices[j]
			}

			got := g.Score("alice", row.remaining)

			if got != nil {
				t.Fatalf("returned error: [%v]", got)
			}
			if got, want := alice.ScoreSheet[Bonus] == 35, row.bonus; got != want {
				t.Errorf("invalid result for scenario %d", i)
			}
		}
	})

	t.Run("should reset roll counter", func(t *testing.T) {
		g := New()
		g.AddPlayer("alice")
		g.Roll("alice")
		g.Roll("alice")

		got := g.Score("alice", Yahtzee)

		if got != nil {
			t.Fatalf("returned error: [%v]", got)
		}
		if got, want := g.RollCount, 0; got != want {
			t.Errorf("log counter is %d instead of %d", got, want)
		}
	})

	t.Run("should switch current to next player", func(t *testing.T) {
		g := New()
		g.AddPlayer("alice")
		g.AddPlayer("alice")
		g.Roll("alice")

		got := g.Score("alice", Chance)

		if got != nil {
			t.Fatalf("returned error: [%v]", got)
		}
		if got, want := g.Current, 1; got != want {
			t.Errorf("current player index is %d instead of %d", got, want)
		}
	})

	t.Run("should set the first player as current after the last one", func(t *testing.T) {
		g := New()
		g.AddPlayer("alice")
		g.AddPlayer("bob")
		g.Roll("alice")
		g.Score("alice", Chance)
		g.Roll("bob")
		got := g.Score("bob", Chance)

		if got != nil {
			t.Fatalf("returned error: [%v]", got)
		}
		if got, want := g.Current, 0; got != want {
			t.Errorf("current player index is %d instead of %d", got, want)
		}
	})

	t.Run("should increment round when first player comes again", func(t *testing.T) {
		g := New()
		g.AddPlayer("alice")
		g.Roll("alice")
		got := g.Score("alice", Chance)

		if got != nil {
			t.Fatalf("returned error: [%v]", got)
		}
		if got, want := g.Round, 1; got != want {
			t.Errorf("round counter is %d instead of %d", got, want)
		}
	})

	t.Run("should unlock all dices", func(t *testing.T) {
		g := New()
		g.AddPlayer("alice")
		g.Roll("alice")
		g.Toggle("alice", 2)
		g.Toggle("alice", 3)
		got := g.Score("alice", Chance)

		if got != nil {
			t.Fatalf("returned error: [%v]", got)
		}
		for i, d := range g.Dices {
			if got, want := d.Locked, false; got != want {
				t.Errorf("dice %d is still locked", i)
			}
		}
	})

	t.Run("should fail when got invalid category", func(t *testing.T) {
		g := New()
		g.AddPlayer("alice")
		g.Roll("alice")

		got := g.Score("alice", Category("fake"))

		if want := ErrInvalidCategory; got != want {
			t.Errorf("wrong result, got %#v wanted %#v.", got, want)
		}
	})

	t.Run("should fail when got bonus category", func(t *testing.T) {
		g := New()
		g.AddPlayer("alice")
		g.Roll("alice")

		got := g.Score("alice", Bonus)

		if want := ErrInvalidCategory; got != want {
			t.Errorf("wrong result, got %#v wanted %#v.", got, want)
		}
	})

	t.Run("should fail when category was already scored", func(t *testing.T) {
		g := New()
		g.AddPlayer("alice")
		g.Roll("alice")
		g.Players[0].ScoreSheet[Twos] = 4

		got := g.Score("alice", Twos)

		if want := ErrCategoryAlreadyScored; got != want {
			t.Errorf("wrong result, got %#v wanted %#v.", got, want)
		}
	})

	t.Run("should fail when game is over", func(t *testing.T) {
		g := New()
		g.AddPlayer("alice")
		g.Round = 13

		got := g.Score("alice", Chance)

		if want := ErrGameOver; got != want {
			t.Errorf("wrong result, got [%#v] wanted [%#v].", got, want)
		}
	})

	t.Run("should fail when there was no roll", func(t *testing.T) {
		g := New()
		g.AddPlayer("alice")

		got := g.Score("alice", Chance)

		if want := ErrNoRollYet; got != want {
			t.Errorf("got [%#v], want [%#v]", got, want)
		}
	})

	t.Run("should return error when not player's turn", func(t *testing.T) {
		g := New()
		g.AddPlayer("alice")
		g.AddPlayer("bob")
		g.Roll("alice")

		got := g.Score("bob", Chance)

		if want := ErrNotPlayersTurn; got != want {
			t.Errorf("wrong result, got %#v wanted %#v.", got, want)
		}
	})
}

func TestGame_Toggle(t *testing.T) {
	t.Run("should lock dice if it was unlocked", func(t *testing.T) {
		g := New()
		g.AddPlayer("alice")
		g.Roll("alice")

		got := g.Toggle("alice", 2)

		if got != nil {
			t.Fatalf("got error [%v]", got)
		}
		if got, want := g.Dices[2].Locked, true; got != want {
			t.Errorf("did not lock dice")
		}
	})

	t.Run("should unlock dice if it was locked", func(t *testing.T) {
		g := New()
		g.AddPlayer("alice")
		g.Roll("alice")
		g.Toggle("alice", 2)

		got := g.Toggle("alice", 2)

		if got != nil {
			t.Fatalf("got error [%v]", got)
		}
		if got, want := g.Dices[2].Locked, false; got != want {
			t.Errorf("did not unlock dice")
		}
	})

	t.Run("should return error when dice index is invalid", func(t *testing.T) {
		table := []struct {
			index int
			want  error
		}{
			{-1, ErrInvalidDice},
			{0, nil},
			{5, ErrInvalidDice},
		}

		for _, scenario := range table {
			g := New()
			g.AddPlayer("alice")
			g.Roll("alice")

			got := g.Toggle("alice", scenario.index)

			if want := scenario.want; got != want {
				t.Errorf("for index [%d] got [%#v] but wanted [%#v]", scenario.index, got, want)
			}
		}
	})

	t.Run("should return error when not player's turn", func(t *testing.T) {
		g := New()
		g.AddPlayer("alice")
		g.AddPlayer("bob")
		g.Roll("alice")

		got := g.Toggle("bob", 1)

		if want := ErrNotPlayersTurn; got != want {
			t.Errorf("wrong result, got %#v wanted %#v.", got, want)
		}
	})

	t.Run("should fail when there was no roll", func(t *testing.T) {
		g := New()
		g.AddPlayer("alice")

		got := g.Toggle("alice", 3)

		if want := ErrNoRollYet; got != want {
			t.Errorf("got [%#v], want [%#v]", got, want)
		}
	})

	t.Run("should return error when no more rolls", func(t *testing.T) {
		g := New()
		g.AddPlayer("alice")
		g.Roll("alice")
		g.Roll("alice")
		g.Roll("alice")

		got := g.Toggle("alice", 4)

		if want := ErrOutOfRolls; got != want {
			t.Errorf("wrong result, got %#v wanted %#v.", got, want)
		}
	})

	t.Run("should return error when game is over", func(t *testing.T) {
		g := New()
		g.AddPlayer("alice")
		g.Roll("alice")
		g.Round = 13

		got := g.Toggle("alice", 3)

		if want := ErrGameOver; got != want {
			t.Errorf("wrong result, got %#v wanted %#v.", got, want)
		}
	})
}
