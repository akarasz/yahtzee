package score

import "github.com/akarasz/yahtzee"

type Scorer interface {
	Score() func(dices []int, scoresheet map[yahtzee.Category]int) int
}

type DefaultOnes struct{}

func (d *DefaultOnes) Score() func(dices []int, scoresheet map[yahtzee.Category]int) int {
	return func(dices []int, scoresheet map[yahtzee.Category]int) int {
		return min(countDice(1, dices), 5*1)
	}
}

type DefaultTwos struct{}

func (d *DefaultTwos) Score() func(dices []int, scoresheet map[yahtzee.Category]int) int {
	return func(dices []int, scoresheet map[yahtzee.Category]int) int {
		return min(countDice(2, dices)*2, 5*2)
	}
}

type DefaultThrees struct{}

func (d *DefaultThrees) Score() func(dices []int, scoresheet map[yahtzee.Category]int) int {
	return func(dices []int, scoresheet map[yahtzee.Category]int) int {
		return min(countDice(3, dices)*2, 5*3)
	}
}

type DefaultFours struct{}

func (d *DefaultFours) Score() func(dices []int, scoresheet map[yahtzee.Category]int) int {
	return func(dices []int, scoresheet map[yahtzee.Category]int) int {
		return min(countDice(4, dices)*2, 5*4)
	}
}

type DefaultFives struct{}

func (d *DefaultFives) Score() func(dices []int, scoresheet map[yahtzee.Category]int) int {
	return func(dices []int, scoresheet map[yahtzee.Category]int) int {
		return min(countDice(5, dices)*5, 5*5)
	}
}

type DefaultSix struct{}

func (d *DefaultSix) Score() func(dices []int, scoresheet map[yahtzee.Category]int) int {
	return func(dices []int, scoresheet map[yahtzee.Category]int) int {
		return min(countDice(6, dices)*6, 5*6)
	}
}

type DefaultThreeOfAKind struct{}

func (d *DefaultThreeOfAKind) Score() func(dices []int, scoresheet map[yahtzee.Category]int) int {
	return func(dices []int, scoresheet map[yahtzee.Category]int) int {
		occurrences := map[int]int{}
		for _, d := range dices {
			occurrences[d]++
		}
		s := 0
		for k, v := range occurrences {
			if v >= 3 {
				s = max(s, 3*k)
			}
		}
		return s
	}
}

type DefaultFourOfAKind struct{}

func (d *DefaultFourOfAKind) Score() func(dices []int, scoresheet map[yahtzee.Category]int) int {
	return func(dices []int, scoresheet map[yahtzee.Category]int) int {
		occurrences := map[int]int{}
		for _, d := range dices {
			occurrences[d]++
		}
		s := 0
		for k, v := range occurrences {
			if v >= 4 {
				s = 4 * k
			}
		}
		return s
	}
}

type DefaultFullHouse struct{}

func (d *DefaultFullHouse) Score() func(dices []int, scoresheet map[yahtzee.Category]int) int {
	return func(dices []int, scoresheet map[yahtzee.Category]int) int {
		occurrences := map[int]int{}
		for _, d := range dices {
			occurrences[d]++
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
}

type DefaultSmallStraight struct{}

func (d *DefaultSmallStraight) Score() func(dices []int, scoresheet map[yahtzee.Category]int) int {
	return func(dices []int, scoresheet map[yahtzee.Category]int) int {
		s := 0
		hit := [6]bool{}
		for _, d := range dices {
			hit[d-1] = true
		}

		if (hit[0] && hit[1] && hit[2] && hit[3]) ||
			(hit[1] && hit[2] && hit[3] && hit[4]) ||
			(hit[2] && hit[3] && hit[4] && hit[5]) {
			s = 30
		}
		return s
	}
}

type DefaultLargeStraight struct{}

func (d *DefaultLargeStraight) Score() func(dices []int, scoresheet map[yahtzee.Category]int) int {
	return func(dices []int, scoresheet map[yahtzee.Category]int) int {
		s := 0
		hit := [6]bool{}
		for _, d := range dices {
			hit[d-1] = true
		}

		if (hit[0] && hit[1] && hit[2] && hit[3] && hit[4]) ||
			(hit[1] && hit[2] && hit[3] && hit[4] && hit[5]) {
			s = 40
		}
		return s
	}
}

type DefaultYahtzee struct{}

func (d *DefaultYahtzee) Score() func(dices []int, scoresheet map[yahtzee.Category]int) int {
	return func(dices []int, scoresheet map[yahtzee.Category]int) int {
		s := 0
		for i := 1; i < 7; i++ {
			sameCount := 0
			for j := 0; j < len(dices); j++ {
				if dices[j] == i {
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
}

type DefaultChance struct{}

func (d *DefaultChance) Score() func(dices []int, scoresheet map[yahtzee.Category]int) int {
	return func(dices []int, scoresheet map[yahtzee.Category]int) int {
		s := 0
		for i := 0; i < len(dices); i++ {
			sum := 0
			for j, d := range dices {
				if len(dices) > 5 && j == i {
					continue
				}
				sum += d
			}
			s = max(s, sum)
		}
		return s
	}
}

func countDice(value int, dices []int) int {
	c := 0
	for _, d := range dices {
		if value == d {
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
