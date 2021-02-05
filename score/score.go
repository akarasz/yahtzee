package score

import "github.com/akarasz/yahtzee"

type Score interface {
	Ones(occurrences map[int]int, scoresheet map[yahtzee.Category]int) int
	Twos(occurrences map[int]int, scoresheet map[yahtzee.Category]int) int
	Threes(occurrences map[int]int, scoresheet map[yahtzee.Category]int) int
	Fours(occurrences map[int]int, scoresheet map[yahtzee.Category]int) int
	Fives(occurrences map[int]int, scoresheet map[yahtzee.Category]int) int
	Sixes(occurrences map[int]int, scoresheet map[yahtzee.Category]int) int
	ThreeOfAKind(occurrences map[int]int, scoresheet map[yahtzee.Category]int) int
	FourOfAKind(occurrences map[int]int, scoresheet map[yahtzee.Category]int) int
	FullHouse(occurrences map[int]int, scoresheet map[yahtzee.Category]int) int
	SmallStraight(dices []int, scoresheet map[yahtzee.Category]int) int
	LargeStraight(dices []int, scoresheet map[yahtzee.Category]int) int
	Yahtzee(dices []int, scoresheet map[yahtzee.Category]int) int
	Chance(dices []int, scoresheet map[yahtzee.Category]int) int
}

type Default struct{}

func (d Default) Ones(occurrences map[int]int, scoresheet map[yahtzee.Category]int) int {
	return min(occurrences[1]*1, 5*1)
}
func (d Default) Twos(occurrences map[int]int, scoresheet map[yahtzee.Category]int) int {
	return min(occurrences[2]*2, 5*2)
}
func (d Default) Threes(occurrences map[int]int, scoresheet map[yahtzee.Category]int) int {
	return min(occurrences[3]*3, 5*3)
}
func (d Default) Fours(occurrences map[int]int, scoresheet map[yahtzee.Category]int) int {
	return min(occurrences[4]*4, 5*4)
}
func (d Default) Fives(occurrences map[int]int, scoresheet map[yahtzee.Category]int) int {
	return min(occurrences[5]*5, 5*5)
}
func (d Default) Sixes(occurrences map[int]int, scoresheet map[yahtzee.Category]int) int {
	return min(occurrences[6]*6, 5*6)
}
func (d Default) ThreeOfAKind(occurrences map[int]int, scoresheet map[yahtzee.Category]int) int {
	s := 0
	for k, v := range occurrences {
		if v >= 3 {
			s = max(s, 3*k)
		}
	}
	return s
}
func (d Default) FourOfAKind(occurrences map[int]int, scoresheet map[yahtzee.Category]int) int {
	s := 0
	for k, v := range occurrences {
		if v >= 4 {
			s = 4 * k
		}
	}
	return s
}
func (d Default) FullHouse(occurrences map[int]int, scoresheet map[yahtzee.Category]int) int {
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
func (d Default) SmallStraight(dices []int, scoresheet map[yahtzee.Category]int) int {
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
func (d Default) LargeStraight(dices []int, scoresheet map[yahtzee.Category]int) int {
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
func (d Default) Yahtzee(dices []int, scoresheet map[yahtzee.Category]int) int {
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
func (d Default) Chance(dices []int, scoresheet map[yahtzee.Category]int) int {
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
