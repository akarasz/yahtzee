package yahtzee

type Score struct {
	PreScoreActions  map[PreScoreAction]func(game *Game)
	ScoreActions     Scorers
	PostScoreActions map[PostScoreAction]func(game *Game)
	PostGameActions  map[PostGameAction]func(game *Game)
}

type PreScoreAction string
type PostScoreAction string
type PostGameAction string

const (
	ChanceBonusAction PostGameAction = "chanceBonusAction"
)

const (
	YahtzeeBonusPreScore PreScoreAction = "yahtzeeBonusPreScoreAction"
)

const (
	YahtzeeBonusPostScore PostScoreAction = "yahtzeeBonusPostScoreAction"
)

type Scorers map[Category]func(game *Game) (int, []PostScoreAction)

var defaultScorer = Scorers{
	Ones:          DefaultOnes,
	Twos:          DefaultTwos,
	Threes:        DefaultThrees,
	Fours:         DefaultFours,
	Fives:         DefaultFives,
	Sixes:         DefaultSixes,
	ThreeOfAKind:  DefaultThreeOfAKind,
	FourOfAKind:   DefaultFourOfAKind,
	FullHouse:     DefaultFullHouse,
	SmallStraight: DefaultSmallStraight,
	LargeStraight: DefaultLargeStraight,
	Yahtzee:       DefaultYahtzee,
	Chance:        DefaultChance,
}

func NewDefaultScorer() Scorers {
	scorer := Scorers{}
	for key, value := range defaultScorer {
		scorer[key] = value
	}
	return scorer
}

func DefaultOnes(game *Game) (int, []PostScoreAction) {
	return min(countDice(1, game.Dices), 5*1), nil
}

func DefaultTwos(game *Game) (int, []PostScoreAction) {
	return min(countDice(2, game.Dices)*2, 5*2), nil
}

func DefaultThrees(game *Game) (int, []PostScoreAction) {
	return min(countDice(3, game.Dices)*3, 5*3), nil
}

func DefaultFours(game *Game) (int, []PostScoreAction) {
	return min(countDice(4, game.Dices)*4, 5*4), nil
}

func DefaultFives(game *Game) (int, []PostScoreAction) {
	return min(countDice(5, game.Dices)*5, 5*5), nil
}

func DefaultSixes(game *Game) (int, []PostScoreAction) {
	return min(countDice(6, game.Dices)*6, 5*6), nil
}

func DefaultThreeOfAKind(game *Game) (int, []PostScoreAction) {
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
	return s, nil
}

func DefaultFourOfAKind(game *Game) (int, []PostScoreAction) {
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
	return s, nil
}

func DefaultFullHouse(game *Game) (int, []PostScoreAction) {
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
	return s, nil
}

func DefaultSmallStraight(game *Game) (int, []PostScoreAction) {
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
	return s, nil
}

func DefaultLargeStraight(game *Game) (int, []PostScoreAction) {
	s := 0
	hit := [6]bool{}
	for _, d := range game.Dices {
		hit[d.Value-1] = true
	}

	if (hit[0] && hit[1] && hit[2] && hit[3] && hit[4]) ||
		(hit[1] && hit[2] && hit[3] && hit[4] && hit[5]) {
		s = 40
	}
	return s, nil
}

func DefaultYahtzee(game *Game) (int, []PostScoreAction) {
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
	return s, nil
}

func DefaultChance(game *Game) (int, []PostScoreAction) {
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
	return s, nil
}

//Yahtzee-bonus

func YahtzeeBonusFullHouse(game *Game) (int, []PostScoreAction) {
	if _, yahtzeeScored := game.Players[game.CurrentPlayer].ScoreSheet[Yahtzee]; yahtzeeScored && isYahtzee(game.Dices) {
		return 25, nil
	}
	return DefaultFullHouse(game)
}

func YahtzeeBonusSmallStraight(game *Game) (int, []PostScoreAction) {
	if _, yahtzeeScored := game.Players[game.CurrentPlayer].ScoreSheet[Yahtzee]; yahtzeeScored && isYahtzee(game.Dices) {
		return 30, nil
	}
	return DefaultSmallStraight(game)
}

func YahtzeeBonusLargeStraight(game *Game) (int, []PostScoreAction) {
	if _, yahtzeeScored := game.Players[game.CurrentPlayer].ScoreSheet[Yahtzee]; yahtzeeScored && isYahtzee(game.Dices) {
		return 40, nil
	}
	return DefaultLargeStraight(game)
}

func YahtzeeBonusPreScoreAction(g *Game) {
	yahtzeeValue, yahtzeeScored := g.Players[g.CurrentPlayer].ScoreSheet[Yahtzee]
	yahtzee := isYahtzee(g.Dices)
	g.Context["yahtzeeBonusEligible"] = yahtzeeScored && yahtzee && yahtzeeValue != 0
}

func YahtzeeBonusPostScoreAction(g *Game) {
	if val, ok := g.Context["yahtzeeBonusEligible"]; ok && val.(bool) {
		g.Players[g.CurrentPlayer].ScoreSheet[Yahtzee] += 100
	}
	delete(g.Context, "yahtzeeBonusEligible")
}

func TheChanceAction(g *Game) {
	for _, p := range g.Players {
		s := 0
		for _, v := range p.ScoreSheet {
			s += v
		}
		if s == 5 {
			p.ScoreSheet[ChanceBonus] = 495
		}
	}
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

func isYahtzee(dices []*Dice) bool {
	for i := 1; i < 7; i++ {
		sameCount := 0
		for j := 0; j < len(dices); j++ {
			if dices[j].Value == i {
				sameCount++
			}
		}
		if sameCount >= 5 {
			return true
		}
	}
	return false
}
