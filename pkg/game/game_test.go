package game

import (
	"testing"

	"github.com/akarasz/yahtzee/pkg/models"
)

func TestAddPlayer(t *testing.T) {
	t.Run("should add player", func(t *testing.T) {
		g := models.NewGame()
		c := New()

		c.AddPlayer(g, "alice")

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
			g := models.NewGame()
			g.CurrentPlayer = row.current
			g.Round = row.round
			c := New()

			got := c.AddPlayer(g, "alice")

			if want := row.expected; got != want {
				t.Errorf("adding to %v was incorrect, got: %#v, want: %#v.", g, got, want)
			}
		}
	})

	t.Run("should fail is player with name is already added", func(t *testing.T) {
		g := models.NewGame()
		g.Players = append(g.Players, &models.Player{"alice", nil})
		c := New()

		got := c.AddPlayer(g, "alice")

		if want := ErrPlayerAlreadyAdded; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestGame_Roll(t *testing.T) {
	t.Run("should return the rolled dices", func(t *testing.T) {
		g := models.NewGame()
		g.Players = append(g.Players, &models.Player{"alice", nil})
		c := New()

		got, err := c.Roll(g, "alice")

		if err != nil {
			t.Fatal(err)
		}
		for i, d := range got {
			if want := g.Dices[i]; d != want {
				t.Errorf("at index %d got %v want %v", i, got, want)
			}
		}
	})

	t.Run("should set valid values for dices", func(t *testing.T) {
		g := models.NewGame()
		g.Players = append(g.Players, &models.Player{"alice", nil})
		for _, d := range g.Dices {
			d.Value = -1
		}
		c := New()

		got, err := c.Roll(g, "alice")

		if err != nil {
			t.Fatalf("returned error: [%v]", got)
		}
		for i, d := range g.Dices {
			if got := d.Value; got < 1 || got > 6 {
				t.Errorf("%dth dice has an invalid value, %d.", i, got)
			}
		}
	})

	t.Run("should not roll locked dices", func(t *testing.T) {
		g := models.NewGame()
		g.Players = append(g.Players, &models.Player{"alice", nil})
		g.Dices[2].Locked = true
		g.Dices[2].Value = -1
		c := New()

		got, err := c.Roll(g, "alice")

		if err != nil {
			t.Fatalf("returned error: [%v]", got)
		}
		if g.Dices[2].Value != -1 {
			t.Errorf("value of locked dice got changed")
		}
	})

	t.Run("should increment roll count", func(t *testing.T) {
		g := models.NewGame()
		g.Players = append(g.Players, &models.Player{"alice", nil})
		c := New()

		got, err := c.Roll(g, "alice")

		if err != nil {
			t.Fatalf("returned error: [%v]", got)
		}
		if want := 1; g.RollCount != want {
			t.Errorf("not got incremented")
		}
	})

	t.Run("should return error when no player in game", func(t *testing.T) {
		g := models.NewGame()
		c := New()

		_, got := c.Roll(g, "alice")

		if want := ErrNotPlayersTurn; got != want {
			t.Errorf("wrong result, got %#v wanted %#v.", got, want)
		}
	})

	t.Run("should return error when not player's turn", func(t *testing.T) {
		g := models.NewGame()
		g.Players = append(g.Players, &models.Player{"alice", nil})
		g.Players = append(g.Players, &models.Player{"bob", nil})
		g.CurrentPlayer = 1
		c := New()

		_, got := c.Roll(g, "alice")

		if want := ErrNotPlayersTurn; got != want {
			t.Errorf("wrong result, got %#v wanted %#v.", got, want)
		}
	})

	t.Run("should return error when out of rolls", func(t *testing.T) {
		g := models.NewGame()
		g.Players = append(g.Players, &models.Player{"alice", nil})
		g.RollCount = 3
		c := New()

		_, got := c.Roll(g, "alice")

		if want := ErrOutOfRolls; got != want {
			t.Errorf("wrong result, got %#v wanted %#v.", got, want)
		}
	})

	t.Run("should return error when out of rounds", func(t *testing.T) {
		g := models.NewGame()
		g.Players = append(g.Players, &models.Player{"alice", nil})
		g.Round = 13
		c := New()

		_, got := c.Roll(g, "alice")

		if want := ErrGameOver; got != want {
			t.Errorf("wrong result, got %#v wanted %#v.", got, want)
		}
	})
}

func TestGame_Score(t *testing.T) {
	t.Run("should calculate points correctly", func(t *testing.T) {
		table := []struct {
			dices    []int
			category models.Category
			value    int
		}{
			{[]int{1, 2, 3, 1, 1}, models.Ones, 3},
			{[]int{2, 3, 4, 2, 3}, models.Twos, 4},
			{[]int{6, 4, 2, 2, 3}, models.Threes, 3},
			{[]int{1, 6, 3, 3, 5}, models.Fours, 0},
			{[]int{4, 4, 1, 2, 4}, models.Fours, 12},
			{[]int{6, 6, 3, 5, 2}, models.Fives, 5},
			{[]int{5, 3, 6, 6, 6}, models.Sixes, 18},
			{[]int{2, 4, 3, 6, 4}, models.ThreeOfAKind, 0},
			{[]int{3, 1, 3, 1, 3}, models.ThreeOfAKind, 9},
			{[]int{5, 2, 5, 5, 5}, models.ThreeOfAKind, 15},
			{[]int{2, 6, 3, 2, 2}, models.FourOfAKind, 0},
			{[]int{1, 6, 6, 6, 6}, models.FourOfAKind, 24},
			{[]int{4, 4, 4, 4, 4}, models.FourOfAKind, 16},
			{[]int{5, 5, 2, 5, 5}, models.FullHouse, 0},
			{[]int{2, 5, 3, 6, 5}, models.FullHouse, 0},
			{[]int{5, 5, 2, 5, 2}, models.FullHouse, 25},
			{[]int{3, 1, 3, 1, 3}, models.FullHouse, 25},
			{[]int{6, 2, 5, 1, 3}, models.SmallStraight, 0},
			{[]int{6, 2, 4, 1, 3}, models.SmallStraight, 30},
			{[]int{4, 2, 3, 5, 3}, models.SmallStraight, 30},
			{[]int{1, 6, 3, 5, 4}, models.SmallStraight, 30},
			{[]int{3, 5, 2, 3, 4}, models.LargeStraight, 0},
			{[]int{3, 5, 2, 1, 4}, models.LargeStraight, 40},
			{[]int{5, 2, 6, 3, 4}, models.LargeStraight, 40},
			{[]int{3, 3, 3, 3, 3}, models.Yahtzee, 50},
			{[]int{1, 1, 1, 1, 1}, models.Yahtzee, 50},
			{[]int{6, 2, 4, 1, 3}, models.Chance, 16},
			{[]int{1, 6, 3, 3, 5}, models.Chance, 18},
			{[]int{2, 3, 4, 2, 3}, models.Chance, 14},
		}

		for _, row := range table {
			g := models.NewGame()
			g.Players = append(g.Players, &models.Player{"alice", map[models.Category]int{}})
			g.RollCount = 1
			for i, v := range row.dices {
				g.Dices[i].Value = v
			}
			c := New()

			got := c.Score(g, "alice", row.category)

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
			remaining models.Category
			dices     []int
			bonus     bool
		}{
			{[]int{3, 6, -1, 16, 25, -1}, models.Sixes, []int{1, 3, 6, 2, 4}, false},
			{[]int{-1, -1, 12, -1, 20, 36}, models.Fours, []int{1, 3, 6, 2, 4}, false},
			{[]int{3, 6, 9, 16, 25, -1}, models.Sixes, []int{1, 3, 6, 2, 4}, true},
			{[]int{-1, 2, 3, 4, 15, 36}, models.Ones, []int{1, 1, 3, 3, 3}, false},
			{[]int{-1, 2, 3, 4, 15, 36}, models.Ones, []int{1, 1, 1, 3, 3}, true},
			{[]int{-1, 2, 3, 4, 15, 36}, models.Ones, []int{1, 1, 1, 1, 3}, true},
		}

		for i, row := range table {
			g := models.NewGame()
			g.RollCount = 1
			alice := &models.Player{
				Name:       "alice",
				ScoreSheet: map[models.Category]int{},
			}
			if row.values[0] > 0 {
				alice.ScoreSheet[models.Ones] = row.values[0]
			}
			if row.values[1] > 0 {
				alice.ScoreSheet[models.Twos] = row.values[1]
			}
			if row.values[2] > 0 {
				alice.ScoreSheet[models.Threes] = row.values[2]
			}
			if row.values[3] > 0 {
				alice.ScoreSheet[models.Fours] = row.values[3]
			}
			if row.values[4] > 0 {
				alice.ScoreSheet[models.Fives] = row.values[4]
			}
			if row.values[5] > 0 {
				alice.ScoreSheet[models.Sixes] = row.values[5]
			}
			g.Players = append(g.Players, alice)
			for j, d := range g.Dices {
				d.Value = row.dices[j]
			}
			c := New()

			got := c.Score(g, "alice", row.remaining)

			if got != nil {
				t.Fatalf("returned error: [%v]", got)
			}
			if got, want := alice.ScoreSheet[models.Bonus] == 35, row.bonus; got != want {
				t.Errorf("invalid result for scenario %d", i)
			}
		}
	})

	t.Run("should reset roll counter", func(t *testing.T) {
		g := models.NewGame()
		g.Players = append(g.Players, &models.Player{"alice", map[models.Category]int{}})
		g.RollCount = 2
		c := New()

		got := c.Score(g, "alice", models.Yahtzee)

		if got != nil {
			t.Fatalf("returned error: [%v]", got)
		}
		if got, want := g.RollCount, 0; got != want {
			t.Errorf("log counter is %d instead of %d", got, want)
		}
	})

	t.Run("should switch current to next player", func(t *testing.T) {
		g := models.NewGame()
		g.Players = append(g.Players, &models.Player{"alice", map[models.Category]int{}})
		g.Players = append(g.Players, &models.Player{"bob", map[models.Category]int{}})
		g.RollCount = 1
		c := New()

		got := c.Score(g, "alice", models.Chance)

		if got != nil {
			t.Fatalf("returned error: [%v]", got)
		}
		if got, want := g.CurrentPlayer, 1; got != want {
			t.Errorf("current player index is %d instead of %d", got, want)
		}
	})

	t.Run("should set the first player as current after the last one", func(t *testing.T) {
		g := models.NewGame()
		g.Players = append(g.Players, &models.Player{"alice", map[models.Category]int{}})
		g.Players = append(g.Players, &models.Player{"bob", map[models.Category]int{}})
		g.CurrentPlayer = 1
		g.RollCount = 1
		c := New()

		got := c.Score(g, "bob", models.Chance)

		if got != nil {
			t.Fatalf("returned error: [%v]", got)
		}
		if got, want := g.CurrentPlayer, 0; got != want {
			t.Errorf("current player index is %d instead of %d", got, want)
		}
	})

	t.Run("should increment round when first player comes again", func(t *testing.T) {
		g := models.NewGame()
		g.Players = append(g.Players, &models.Player{"alice", map[models.Category]int{}})
		g.RollCount = 1
		c := New()

		got := c.Score(g, "alice", models.Chance)

		if got != nil {
			t.Fatalf("returned error: [%v]", got)
		}
		if got, want := g.Round, 1; got != want {
			t.Errorf("round counter is %d instead of %d", got, want)
		}
	})

	t.Run("should unlock all dices", func(t *testing.T) {
		g := models.NewGame()
		g.Players = append(g.Players, &models.Player{"alice", map[models.Category]int{}})
		g.RollCount = 1
		g.Dices[2].Locked = true
		g.Dices[3].Locked = true
		c := New()

		got := c.Score(g, "alice", models.Chance)

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
		g := models.NewGame()
		g.Players = append(g.Players, &models.Player{"alice", map[models.Category]int{}})
		g.RollCount = 1
		c := New()

		got := c.Score(g, "alice", models.Category("fake"))

		if want := ErrInvalidCategory; got != want {
			t.Errorf("wrong result, got %#v wanted %#v.", got, want)
		}
	})

	t.Run("should fail when got bonus category", func(t *testing.T) {
		g := models.NewGame()
		g.Players = append(g.Players, &models.Player{"alice", map[models.Category]int{}})
		g.RollCount = 1
		c := New()

		got := c.Score(g, "alice", models.Bonus)

		if want := ErrInvalidCategory; got != want {
			t.Errorf("wrong result, got %#v wanted %#v.", got, want)
		}
	})

	t.Run("should fail when category was already scored", func(t *testing.T) {
		g := models.NewGame()
		g.Players = append(g.Players, &models.Player{"alice", map[models.Category]int{
			models.Twos: 4,
		}})
		g.RollCount = 1
		c := New()

		got := c.Score(g, "alice", models.Twos)

		if want := ErrCategoryAlreadyScored; got != want {
			t.Errorf("wrong result, got %#v wanted %#v.", got, want)
		}
	})

	t.Run("should fail when game is over", func(t *testing.T) {
		g := models.NewGame()
		g.Players = append(g.Players, &models.Player{"alice", map[models.Category]int{}})
		g.RollCount = 1
		g.Round = 13
		c := New()

		got := c.Score(g, "alice", models.Chance)

		if want := ErrGameOver; got != want {
			t.Errorf("wrong result, got [%#v] wanted [%#v].", got, want)
		}
	})

	t.Run("should fail when there was no roll", func(t *testing.T) {
		g := models.NewGame()
		g.Players = append(g.Players, &models.Player{"alice", map[models.Category]int{}})
		c := New()

		got := c.Score(g, "alice", models.Chance)

		if want := ErrNoRollYet; got != want {
			t.Errorf("got [%#v], want [%#v]", got, want)
		}
	})

	t.Run("should return error when not player's turn", func(t *testing.T) {
		g := models.NewGame()
		g.Players = append(g.Players, &models.Player{"alice", map[models.Category]int{}})
		g.Players = append(g.Players, &models.Player{"bob", map[models.Category]int{}})
		g.RollCount = 1
		c := New()

		got := c.Score(g, "bob", models.Chance)

		if want := ErrNotPlayersTurn; got != want {
			t.Errorf("wrong result, got %#v wanted %#v.", got, want)
		}
	})

	t.Run("should return error when no player is in the game", func(t *testing.T) {
		g := models.NewGame()
		c := New()

		got := c.Score(g, "bob", models.Chance)

		if want := ErrNotPlayersTurn; got != want {
			t.Errorf("wrong result, got %#v wanted %#v.", got, want)
		}
	})
}

func TestGame_Toggle(t *testing.T) {
	t.Run("should return the dices", func(t *testing.T) {
		g := models.NewGame()
		g.Players = append(g.Players, &models.Player{"alice", map[models.Category]int{}})
		g.RollCount = 1
		c := New()

		got, err := c.Toggle(g, "alice", 2)

		if err != nil {
			t.Fatal(err)
		}
		for i, d := range got {
			if want := g.Dices[i]; d != want {
				t.Errorf("at index %d got %v want %v", i, got, want)
			}
		}
	})

	t.Run("should lock dice if it was unlocked", func(t *testing.T) {
		g := models.NewGame()
		g.Players = append(g.Players, &models.Player{"alice", map[models.Category]int{}})
		g.RollCount = 1
		c := New()

		got, err := c.Toggle(g, "alice", 2)

		if err != nil {
			t.Fatalf("got error [%v]", got)
		}
		if got, want := g.Dices[2].Locked, true; got != want {
			t.Errorf("did not lock dice")
		}
	})

	t.Run("should unlock dice if it was locked", func(t *testing.T) {
		g := models.NewGame()
		g.Players = append(g.Players, &models.Player{"alice", map[models.Category]int{}})
		g.RollCount = 1
		g.Dices[2].Locked = true
		c := New()

		got, err := c.Toggle(g, "alice", 2)

		if err != nil {
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
			g := models.NewGame()
			g.Players = append(g.Players, &models.Player{"alice", map[models.Category]int{}})
			g.RollCount = 1
			c := New()

			_, got := c.Toggle(g, "alice", scenario.index)

			if want := scenario.want; got != want {
				t.Errorf("for index [%d] got [%#v] but wanted [%#v]", scenario.index, got, want)
			}
		}
	})

	t.Run("should return error when not player's turn", func(t *testing.T) {
		g := models.NewGame()
		g.Players = append(g.Players, &models.Player{"alice", map[models.Category]int{}})
		g.Players = append(g.Players, &models.Player{"bob", map[models.Category]int{}})
		g.RollCount = 1
		c := New()

		_, got := c.Toggle(g, "bob", 1)

		if want := ErrNotPlayersTurn; got != want {
			t.Errorf("wrong result, got %#v wanted %#v.", got, want)
		}
	})

	t.Run("should return error when no player was added to the game", func(t *testing.T) {
		g := models.NewGame()
		c := New()

		_, got := c.Toggle(g, "bob", 1)

		if want := ErrNotPlayersTurn; got != want {
			t.Errorf("wrong result, got %#v wanted %#v.", got, want)
		}
	})

	t.Run("should fail when there was no roll", func(t *testing.T) {
		g := models.NewGame()
		g.Players = append(g.Players, &models.Player{"alice", map[models.Category]int{}})
		c := New()

		_, got := c.Toggle(g, "alice", 3)

		if want := ErrNoRollYet; got != want {
			t.Errorf("got [%#v], want [%#v]", got, want)
		}
	})

	t.Run("should return error when no more rolls", func(t *testing.T) {
		g := models.NewGame()
		g.Players = append(g.Players, &models.Player{"alice", map[models.Category]int{}})
		g.RollCount = 3
		c := New()

		_, got := c.Toggle(g, "alice", 4)

		if want := ErrOutOfRolls; got != want {
			t.Errorf("wrong result, got %#v wanted %#v.", got, want)
		}
	})

	t.Run("should return error when game is over", func(t *testing.T) {
		g := models.NewGame()
		g.Players = append(g.Players, &models.Player{"alice", map[models.Category]int{}})
		g.RollCount = 1
		g.Round = 13
		c := New()

		_, got := c.Toggle(g, "alice", 3)

		if want := ErrGameOver; got != want {
			t.Errorf("wrong result, got %#v wanted %#v.", got, want)
		}
	})
}
