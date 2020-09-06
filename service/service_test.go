package service

import (
	"testing"

	"github.com/akarasz/yahtzee/models"
)

func serviceWithEmptyGameAndSingeUser(name string) *Default {
	u := models.User(name)
	result := &Default{
		game: *models.NewGame(),
		user: u,
	}
	result.game.Players = append(result.game.Players, models.NewPlayer(u))
	return result
}

func TestDefault_AddPlayer(t *testing.T) {
	t.Run("should add player", func(t *testing.T) {
		want := models.User("Alice")
		s := Default{
			game: *models.NewGame(),
			user: want,
		}

		got, err := s.AddPlayer()

		if err != nil {
			t.Fatalf("unexpected error: %T: %v", err, err)
		}
		if len(got.Players) != 1 {
			t.Fatalf("player was not added")
		}
		if got := got.Players[0].User; got != want {
			t.Errorf("not the expected player; got %q, want %q", got, want)
		}
	})

	t.Run("should fail after game started", func(t *testing.T) {
		table := []struct {
			current, round int
			expected       error
			desc           string
		}{
			{0, 0, nil, "not started yet"},
			{0, 1, ErrAlreadyStarted, "first user and second round"},
			{1, 0, ErrAlreadyStarted, "second user and first round"},
			{2, 3, ErrAlreadyStarted, "third user and fourth round"},
		}

		for _, scenario := range table {
			s := Default{
				game: *models.NewGame(),
				user: models.User("Alice"),
			}
			s.game.CurrentPlayer = scenario.current
			s.game.Round = scenario.round

			_, got := s.AddPlayer()
			if want := scenario.expected; got != want {
				t.Errorf("error for %q was incorrect; got: %T, want: %T.", scenario.desc, got, want)
			}
		}
	})

	t.Run("should fail is player with name is already added", func(t *testing.T) {
		s := Default{
			game: *models.NewGame(),
			user: models.User("Alice"),
		}
		s.game.Players = append(s.game.Players, models.NewPlayer(models.User("Alice")))

		_, got := s.AddPlayer()

		if want := ErrPlayerAlreadyAdded; got != want {
			t.Errorf("unexpected error; got %T, want %T", got, want)
		}
	})
}

func TestDefault_Roll(t *testing.T) {
	t.Run("should set valid values for dices", func(t *testing.T) {
		s := serviceWithEmptyGameAndSingeUser("Alice")
		for _, d := range s.game.Dices {
			d.Value = -1
		}

		got, err := s.Roll()

		if err != nil {
			t.Fatalf("unexpected error: %T: %v", err, err)
		}
		for i, d := range got.Dices {
			if got := d.Value; got < 1 || got > 6 {
				t.Errorf("%dth dice has an invalid value: %d.", i, got)
			}
		}
	})

	t.Run("should not roll locked dices", func(t *testing.T) {
		s := serviceWithEmptyGameAndSingeUser("Alice")
		s.game.Dices[2].Locked = true
		s.game.Dices[2].Value = -1

		got, err := s.Roll()

		if err != nil {
			t.Fatalf("unexpected error: %T: %v", err, err)
		}
		if got.Dices[2].Value != -1 {
			t.Errorf("value of locked dice got changed")
		}
	})

	t.Run("should increment roll count", func(t *testing.T) {
		s := serviceWithEmptyGameAndSingeUser("Alice")

		got, err := s.Roll()

		if err != nil {
			t.Fatalf("unexpected error: %T: %v", err, err)
		}
		if want := s.game.RollCount + 1; got.RollCount != want {
			t.Errorf("not got incremented - %d %d", s.game.RollCount, got.RollCount)
		}
	})

	t.Run("should return error when no player in game", func(t *testing.T) {
		s := Default{
			game: *models.NewGame(),
			user: models.User("Alice"),
		}

		_, got := s.Roll()

		if want := ErrNotPlayersTurn; got != want {
			t.Errorf("unexpected error; got %T want %T", got, want)
		}
	})

	t.Run("should return error when not player's turn", func(t *testing.T) {
		s := serviceWithEmptyGameAndSingeUser("Alice")
		s.game.Players = append(s.game.Players, models.NewPlayer(models.User("Bob")))
		s.game.CurrentPlayer = 1

		_, got := s.Roll()

		if want := ErrNotPlayersTurn; got != want {
			t.Errorf("unexpected error; got %T want %T", got, want)
		}
	})

	t.Run("should return error when out of rolls", func(t *testing.T) {
		s := serviceWithEmptyGameAndSingeUser("Alice")
		s.game.RollCount = 3

		_, got := s.Roll()

		if want := ErrOutOfRolls; got != want {
			t.Errorf("unexpected error; got %T want %T", got, want)
		}
	})

	t.Run("should return error when out of rounds", func(t *testing.T) {
		s := serviceWithEmptyGameAndSingeUser("Alice")
		s.game.Round = 13

		_, got := s.Roll()

		if want := ErrGameOver; got != want {
			t.Errorf("unexpected error; got %T want %T", got, want)
		}
	})
}

func TestDefault_Lock(t *testing.T) {
	t.Run("should lock dice if it was unlocked", func(t *testing.T) {
		s := serviceWithEmptyGameAndSingeUser("Alice")
		s.game.RollCount = 1

		got, err := s.Lock(2)

		if err != nil {
			t.Fatalf("unexpected error %T: %v", err, err)
		}
		if got, want := got.Dices[2].Locked, true; got != want {
			t.Errorf("did not lock dice")
		}
	})

	t.Run("should unlock dice if it was locked", func(t *testing.T) {
		s := serviceWithEmptyGameAndSingeUser("Alice")
		s.game.RollCount = 1
		s.game.Dices[2].Locked = true

		got, err := s.Lock(2)

		if err != nil {
			t.Fatalf("unexpected error %T: %v", err, err)
		}
		if got, want := got.Dices[2].Locked, false; got != want {
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
			s := serviceWithEmptyGameAndSingeUser("Alice")
			s.game.RollCount = 1

			_, got := s.Lock(scenario.index)

			if want := scenario.want; got != want {
				t.Errorf("unexpected error for dice index %d; got %T want %T", scenario.index, got, want)
			}
		}
	})

	t.Run("should return error when not player's turn", func(t *testing.T) {
		s := serviceWithEmptyGameAndSingeUser("Alice")
		s.game.Players = append(s.game.Players, models.NewPlayer(models.User("Bob")))
		s.game.CurrentPlayer = 1
		s.game.RollCount = 1

		_, got := s.Lock(1)

		if want := ErrNotPlayersTurn; got != want {
			t.Errorf("unexpected error; got %T want %T", got, want)
		}
	})

	t.Run("should return error when no player was added to the game", func(t *testing.T) {
		s := Default{
			game: *models.NewGame(),
			user: models.User("Alice"),
		}
		s.game.RollCount = 1

		_, got := s.Lock(1)

		if want := ErrNotPlayersTurn; got != want {
			t.Errorf("unexpected error; got %T want %T", got, want)
		}
	})

	t.Run("should fail when there was no roll", func(t *testing.T) {
		s := serviceWithEmptyGameAndSingeUser("Alice")

		_, got := s.Lock(1)

		if want := ErrNoRollYet; got != want {
			t.Errorf("unexpected error; got %T want %T", got, want)
		}
	})

	t.Run("should return error when no more rolls", func(t *testing.T) {
		s := serviceWithEmptyGameAndSingeUser("Alice")
		s.game.RollCount = 3

		_, got := s.Lock(1)

		if want := ErrOutOfRolls; got != want {
			t.Errorf("unexpected error; got %T want %T", got, want)
		}
	})

	t.Run("should return error when game is over", func(t *testing.T) {
		s := serviceWithEmptyGameAndSingeUser("Alice")
		s.game.RollCount = 1
		s.game.Round = 13

		_, got := s.Lock(1)

		if want := ErrGameOver; got != want {
			t.Errorf("unexpected error; got %T want %T", got, want)
		}
	})
}

func TestDefault_Score(t *testing.T) {
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

		for i, scenario := range table {
			s := serviceWithEmptyGameAndSingeUser("Alice")
			s.game.RollCount = 1
			if scenario.values[0] > 0 {
				s.game.Players[0].ScoreSheet[models.Ones] = scenario.values[0]
			}
			if scenario.values[1] > 0 {
				s.game.Players[0].ScoreSheet[models.Twos] = scenario.values[1]
			}
			if scenario.values[2] > 0 {
				s.game.Players[0].ScoreSheet[models.Threes] = scenario.values[2]
			}
			if scenario.values[3] > 0 {
				s.game.Players[0].ScoreSheet[models.Fours] = scenario.values[3]
			}
			if scenario.values[4] > 0 {
				s.game.Players[0].ScoreSheet[models.Fives] = scenario.values[4]
			}
			if scenario.values[5] > 0 {
				s.game.Players[0].ScoreSheet[models.Sixes] = scenario.values[5]
			}
			for j, d := range s.game.Dices {
				d.Value = scenario.dices[j]
			}

			got, err := s.Score(scenario.remaining)

			if err != nil {
				t.Fatalf("unexpected error: %T: %v", err, err)
			}
			if got, want := got.Players[0].ScoreSheet[models.Bonus] == 35, scenario.bonus; got != want {
				t.Errorf("invalid result for scenario %d", i)
			}
		}
	})

	t.Run("should reset roll counter", func(t *testing.T) {
		s := serviceWithEmptyGameAndSingeUser("Alice")
		s.game.RollCount = 2

		got, err := s.Score(models.Yahtzee)

		if err != nil {
			t.Fatalf("unexpected error: %T: %v", err, err)
		}
		if got, want := got.RollCount, 0; got != want {
			t.Errorf("wrong RollCount; got %d want %d", got, want)
		}
	})

	t.Run("should switch current to next player", func(t *testing.T) {
		s := serviceWithEmptyGameAndSingeUser("Alice")
		s.game.Players = append(s.game.Players, models.NewPlayer(models.User("Bob")))
		s.game.RollCount = 1

		got, err := s.Score(models.Chance)

		if err != nil {
			t.Fatalf("unexpected error: %T: %v", err, err)
		}
		if got, want := got.CurrentPlayer, 1; got != want {
			t.Errorf("wrong CurrentUser; got %d want %d", got, want)
		}
	})

	t.Run("should set the first player as current after the last one", func(t *testing.T) {
		bob := models.User("Bob")
		s := serviceWithEmptyGameAndSingeUser("Alice")
		s.game.Players = append(s.game.Players, models.NewPlayer(bob))
		s.game.RollCount = 1
		s.game.CurrentPlayer = 1
		s.user = bob

		got, err := s.Score(models.Chance)

		if err != nil {
			t.Fatalf("unexpected error: %T: %v", err, err)
		}
		if got, want := got.CurrentPlayer, 0; got != want {
			t.Errorf("wrong CurrentUser; got %d want %d", got, want)
		}
	})

	t.Run("should increment round when first player comes again", func(t *testing.T) {
		s := serviceWithEmptyGameAndSingeUser("Alice")
		s.game.RollCount = 1

		got, err := s.Score(models.Chance)

		if err != nil {
			t.Fatalf("unexpected error: %T: %v", err, err)
		}
		if got, want := got.Round, 1; got != want {
			t.Errorf("wrong Round; got %d want %d", got, want)
		}
	})

	t.Run("should unlock all dices", func(t *testing.T) {
		s := serviceWithEmptyGameAndSingeUser("Alice")
		s.game.RollCount = 1
		s.game.Dices[2].Locked = true
		s.game.Dices[3].Locked = true

		got, err := s.Score(models.Chance)

		if err != nil {
			t.Fatalf("unexpected error: %T: %v", err, err)
		}
		for i, d := range got.Dices {
			if got, want := d.Locked, false; got != want {
				t.Errorf("dice %d is still locked", i)
			}
		}
	})

	t.Run("should fail when got invalid category", func(t *testing.T) {
		s := serviceWithEmptyGameAndSingeUser("Alice")
		s.game.RollCount = 1

		_, got := s.Score(models.Category("fake"))

		if want := ErrInvalidCategory; got != want {
			t.Errorf("unexpected error; got %T want %T", got, want)
		}
	})

	t.Run("should fail when got bonus category", func(t *testing.T) {
		s := serviceWithEmptyGameAndSingeUser("Alice")
		s.game.RollCount = 1

		_, got := s.Score(models.Bonus)

		if want := ErrInvalidCategory; got != want {
			t.Errorf("unexpected error; got %T want %T", got, want)
		}
	})

	t.Run("should fail when category was already scored", func(t *testing.T) {
		s := serviceWithEmptyGameAndSingeUser("Alice")
		s.game.Players[0].ScoreSheet[models.Twos] = 0
		s.game.RollCount = 1

		_, got := s.Score(models.Twos)

		if want := ErrCategoryAlreadyScored; got != want {
			t.Errorf("unexpected error; got %T want %T", got, want)
		}
	})

	t.Run("should fail when game is over", func(t *testing.T) {
		s := serviceWithEmptyGameAndSingeUser("Alice")
		s.game.RollCount = 1
		s.game.Round = 13

		_, got := s.Score(models.Twos)

		if want := ErrGameOver; got != want {
			t.Errorf("unexpected error; got %T want %T", got, want)
		}
	})

	t.Run("should fail when there was no roll", func(t *testing.T) {
		s := serviceWithEmptyGameAndSingeUser("Alice")

		_, got := s.Score(models.Chance)

		if want := ErrNoRollYet; got != want {
			t.Errorf("unexpected error; got %T want %T", got, want)
		}
	})

	t.Run("should return error when not player's turn", func(t *testing.T) {
		s := serviceWithEmptyGameAndSingeUser("Alice")
		s.game.Players = append(s.game.Players, models.NewPlayer(models.User("Bob")))
		s.game.RollCount = 1
		s.game.CurrentPlayer = 1

		_, got := s.Score(models.Chance)

		if want := ErrNotPlayersTurn; got != want {
			t.Errorf("unexpected error; got %T want %T", got, want)
		}
	})

	t.Run("should return error when no player is in the game", func(t *testing.T) {
		s := Default{
			game: *models.NewGame(),
			user: models.User("Alice"),
		}

		_, got := s.Score(models.Chance)

		if want := ErrNotPlayersTurn; got != want {
			t.Errorf("unexpected error; got %T want %T", got, want)
		}
	})
}

func TestScore(t *testing.T) {
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
			got, err := Score(row.category, row.dices)

			if err != nil {
				t.Fatalf("returned error: [%v]", got)
			}
			if want := row.value; got != want {
				t.Errorf("%q score for [%v] should be %d but was %d.",
					row.category,
					row.dices,
					want,
					got)
			}
		}
	})

}
