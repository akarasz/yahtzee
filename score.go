package yahtzee

type Score struct {
	PreScoreActions  []func(game *Game)
	ScoreActions     Scorers
	PostScoreActions []func(game *Game)
	PostGameActions  []func(game *Game)
}

type Scorers map[Category]func(game *Game) int

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

func DefaultOnes(game *Game) int {
	return min(countDice(1, game.Dices), 5*1)
}

func DefaultTwos(game *Game) int {
	return min(countDice(2, game.Dices)*2, 5*2)
}

func DefaultThrees(game *Game) int {
	return min(countDice(3, game.Dices)*3, 5*3)
}

func DefaultFours(game *Game) int {
	return min(countDice(4, game.Dices)*4, 5*4)
}

func DefaultFives(game *Game) int {
	return min(countDice(5, game.Dices)*5, 5*5)
}

func DefaultSixes(game *Game) int {
	return min(countDice(6, game.Dices)*6, 5*6)
}

func DefaultThreeOfAKind(game *Game) int {
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

func DefaultFourOfAKind(game *Game) int {
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

func DefaultFullHouse(game *Game) int {
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

func DefaultSmallStraight(game *Game) int {
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

func DefaultLargeStraight(game *Game) int {
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

func DefaultYahtzee(game *Game) int {
	if isYahtzee(game.Dices) {
		return 50
	}
	return 0
}

func DefaultChance(game *Game) int {
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

func DefaultUpperSectionBonusAction(game *Game) {
	if _, ok := game.Players[game.CurrentPlayer].ScoreSheet[Bonus]; !ok {
		var total, types int
		for k, v := range game.Players[game.CurrentPlayer].ScoreSheet {
			if k == Ones || k == Twos || k == Threes ||
				k == Fours || k == Fives || k == Sixes {
				types++
				total += v
			}
		}

		if total >= 63 {
			game.Players[game.CurrentPlayer].ScoreSheet[Bonus] = 35
		} else if types == 6 {
			game.Players[game.CurrentPlayer].ScoreSheet[Bonus] = 0
		}
	}
}

//Yahtzee-bonus

func YahtzeeBonusFullHouse(game *Game) int {
	if _, yahtzeeScored := game.Players[game.CurrentPlayer].ScoreSheet[Yahtzee]; yahtzeeScored && isYahtzee(game.Dices) {
		return 25
	}
	return DefaultFullHouse(game)
}

func YahtzeeBonusSmallStraight(game *Game) int {
	if _, yahtzeeScored := game.Players[game.CurrentPlayer].ScoreSheet[Yahtzee]; yahtzeeScored && isYahtzee(game.Dices) {
		return 30
	}
	return DefaultSmallStraight(game)
}

func YahtzeeBonusLargeStraight(game *Game) int {
	if _, yahtzeeScored := game.Players[game.CurrentPlayer].ScoreSheet[Yahtzee]; yahtzeeScored && isYahtzee(game.Dices) {
		return 40
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

// The Chance

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

// Equilizer

func EquilizerPreScoreAction(g *Game) {
	notScoredCategories := []Category{}
	for _, c := range Categories() {
		if _, ok := g.Players[g.CurrentPlayer].ScoreSheet[c]; !ok {
			notScoredCategories = append(notScoredCategories, c)
		}
	}

	g.Context["equilizerNotScoredCategories"] = notScoredCategories
}

func EquilizerPostScoreAction(g *Game) {
	defer delete(g.Context, "equilizerNotScoredCategories")

	if val, ok := g.Context["equilizerNotScoredCategories"]; ok {
		for _, c := range val.([]Category) {
			if s, ok := g.Players[g.CurrentPlayer].ScoreSheet[c]; ok {
				if s > 0 {
					return
				}
				for _, p := range g.Players {
					if _, ok := p.ScoreSheet[c]; ok {
						p.ScoreSheet[c] = 0
					}
				}
				return
			}
		}
	}
}

// Official

func OfficialThreeOfAKind(game *Game) int {
	occurrences := map[int]int{}
	for _, d := range game.Dices {
		occurrences[d.Value]++
	}
	for _, v := range occurrences {
		if v >= 3 {
			return DefaultChance(game)
		}
	}
	return 0
}

func OfficialFourOfAKind(game *Game) int {
	occurrences := map[int]int{}
	for _, d := range game.Dices {
		occurrences[d.Value]++
	}
	for _, v := range occurrences {
		if v >= 4 {
			return DefaultChance(game)
		}
	}
	return 0
}

func OfficialFullHouse(game *Game) int {
	if _, yahtzeeScored := game.Players[game.CurrentPlayer].ScoreSheet[Yahtzee]; yahtzeeScored && isYahtzee(game.Dices) {
		occurrences := map[int]int{}
		for _, d := range game.Dices {
			occurrences[d.Value]++
		}
		for v, c := range occurrences {
			if _, scored := game.Players[game.CurrentPlayer].ScoreSheet[Categories()[v-1]]; scored && c >= 5 {
				return 25
			}
		}
	}
	return DefaultFullHouse(game)
}

func OfficialSmallStraight(game *Game) int {
	if _, yahtzeeScored := game.Players[game.CurrentPlayer].ScoreSheet[Yahtzee]; yahtzeeScored && isYahtzee(game.Dices) {
		occurrences := map[int]int{}
		for _, d := range game.Dices {
			occurrences[d.Value]++
		}
		for v, c := range occurrences {
			if _, scored := game.Players[game.CurrentPlayer].ScoreSheet[Categories()[v-1]]; scored && c >= 5 {
				return 30
			}
		}
	}
	return DefaultSmallStraight(game)
}

func OfficialLargeStraight(game *Game) int {
	if _, yahtzeeScored := game.Players[game.CurrentPlayer].ScoreSheet[Yahtzee]; yahtzeeScored && isYahtzee(game.Dices) {
		occurrences := map[int]int{}
		for _, d := range game.Dices {
			occurrences[d.Value]++
		}
		for v, c := range occurrences {
			if _, scored := game.Players[game.CurrentPlayer].ScoreSheet[Categories()[v-1]]; scored && c >= 5 {
				return 40
			}
		}
	}
	return DefaultLargeStraight(game)
}

func OfficialYahtzeeBonusPreScoreAction(game *Game) {
	YahtzeeBonusPreScoreAction(game)
}

func OfficialYahtzeeBonusPostScoreAction(game *Game) {
	YahtzeeBonusPostScoreAction(game)
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
