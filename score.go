package yahtzee

type Scorer interface {
	Score(game *Game) int
}

type Scorers map[Category]Scorer

var DefaultScorer = Scorers{
	Ones:          &DefaultOnes{},
	Twos:          &DefaultTwos{},
	Threes:        &DefaultThrees{},
	Fours:         &DefaultFours{},
	Fives:         &DefaultFives{},
	Sixes:         &DefaultSixes{},
	ThreeOfAKind:  &DefaultThreeOfAKind{},
	FourOfAKind:   &DefaultFourOfAKind{},
	FullHouse:     &DefaultFullHouse{},
	SmallStraight: &DefaultSmallStraight{},
	LargeStraight: &DefaultLargeStraight{},
	Yahtzee:       &DefaultYahtzee{},
	Chance:        &DefaultChance{},
}

type DefaultOnes struct{}

func (d *DefaultOnes) Score(game *Game) int {
	return min(countDice(1, game.Dices), 5*1)
}

type DefaultTwos struct{}

func (d *DefaultTwos) Score(game *Game) int {
	return min(countDice(2, game.Dices)*2, 5*2)
}

type DefaultThrees struct{}

func (d *DefaultThrees) Score(game *Game) int {
	return min(countDice(3, game.Dices)*3, 5*3)
}

type DefaultFours struct{}

func (d *DefaultFours) Score(game *Game) int {
	return min(countDice(4, game.Dices)*4, 5*4)
}

type DefaultFives struct{}

func (d *DefaultFives) Score(game *Game) int {
	return min(countDice(5, game.Dices)*5, 5*5)
}

type DefaultSixes struct{}

func (d *DefaultSixes) Score(game *Game) int {
	return min(countDice(6, game.Dices)*6, 5*6)
}

type DefaultThreeOfAKind struct{}

func (d *DefaultThreeOfAKind) Score(game *Game) int {
	occurrences := map[int]int{}
	for _, d := range game.Dices {
		occurrences[d.Value]++
	}
	s := 0
	for k, v := range occurrences {
		if v >= 3 {
			s = max(s, 3*k)
		}
	}
	return s
}

type DefaultFourOfAKind struct{}

func (d *DefaultFourOfAKind) Score(game *Game) int {
	occurrences := map[int]int{}
	for _, d := range game.Dices {
		occurrences[d.Value]++
	}
	s := 0
	for k, v := range occurrences {
		if v >= 4 {
			s = 4 * k
		}
	}
	return s
}

type DefaultFullHouse struct{}

func (d *DefaultFullHouse) Score(game *Game) int {
	occurrences := map[int]int{}
	for _, d := range game.Dices {
		occurrences[d.Value]++
	}
	s := 0
	three := false
	two := false
	for _, v := range occurrences {
		if !three && v >= 3 {
			three = true
			continue
		}
		if !two && v >= 2 {
			two = true
			continue
		}
	}

	if three && two {
		s = 25
	}
	return s
}

type DefaultSmallStraight struct{}

func (d *DefaultSmallStraight) Score(game *Game) int {
	s := 0
	hit := [6]bool{}
	for _, d := range game.Dices {
		hit[d.Value-1] = true
	}

	if (hit[0] && hit[1] && hit[2] && hit[3]) ||
		(hit[1] && hit[2] && hit[3] && hit[4]) ||
		(hit[2] && hit[3] && hit[4] && hit[5]) {
		s = 30
	}
	return s
}

type DefaultLargeStraight struct{}

func (d *DefaultLargeStraight) Score(game *Game) int {
	s := 0
	hit := [6]bool{}
	for _, d := range game.Dices {
		hit[d.Value-1] = true
	}

	if (hit[0] && hit[1] && hit[2] && hit[3] && hit[4]) ||
		(hit[1] && hit[2] && hit[3] && hit[4] && hit[5]) {
		s = 40
	}
	return s
}

type DefaultYahtzee struct{}

func (d *DefaultYahtzee) Score(game *Game) int {
	s := 0
	for i := 1; i < 7; i++ {
		sameCount := 0
		for j := 0; j < len(game.Dices); j++ {
			if game.Dices[j].Value == i {
				sameCount++
			}
		}
		if sameCount >= 5 {
			s = 50
			break
		}
	}
	return s
}

type DefaultChance struct{}

func (d *DefaultChance) Score(game *Game) int {
	s := 0
	for i := 0; i < len(game.Dices); i++ {
		sum := 0
		for j, d := range game.Dices {
			if len(game.Dices) > 5 && j == i {
				continue
			}
			sum += d.Value
		}
		s = max(s, sum)
	}
	return s
}

type DefaultChanceBonus struct{}

func (d *DefaultChanceBonus) Score(game *Game) int {
	if len(game.Players[game.CurrentPlayer].ScoreSheet) >= len(game.Scorer)-1 {
		s := 0
		for _, v := range game.Players[game.CurrentPlayer].ScoreSheet {
			s += v
		}
		if s == 5 {
			return 495
		}
	}
	return 0
}

func countDice(value int, dices []*Dice) int {
	c := 0
	for _, d := range dices {
		if value == d.Value {
			c++
		}
	}
	return c
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
