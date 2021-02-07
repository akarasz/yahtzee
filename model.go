package yahtzee

var (
	// NumberOfDices shows how many dices are used for a game.
	NumberOfDices int = 5
)

// Dice represents a dice you use for the Game.
type Dice struct {
	// Value is the number on the face of the dice
	Value int

	// Locked shows if the dice will roll or not
	Locked bool
}

// Category represents the formations players try to roll.
type Category string

// Available categories
const (
	Ones   Category = "ones"
	Twos   Category = "twos"
	Threes Category = "threes"
	Fours  Category = "fours"
	Fives  Category = "fives"
	Sixes  Category = "sixes"
	Bonus  Category = "bonus"

	ThreeOfAKind  Category = "three-of-a-kind"
	FourOfAKind   Category = "four-of-a-kind"
	FullHouse     Category = "full-house"
	SmallStraight Category = "small-straight"
	LargeStraight Category = "large-straight"
	Yahtzee       Category = "yahtzee"
	Chance        Category = "chance"
	ChanceBonus   Category = "chance-bonus"
)

func Categories() []Category {
	return []Category{
		Ones,
		Twos,
		Threes,
		Fours,
		Fives,
		Sixes,
		ThreeOfAKind,
		FourOfAKind,
		FullHouse,
		SmallStraight,
		LargeStraight,
		Yahtzee,
		Chance,
	}
}

// Player contains all data representing a player.
type Player struct {
	// User who plays
	User User

	// ScoreSheet keeps the scores of the player
	ScoreSheet map[Category]int
}

// NewPlayer returns a new named player with an empty score sheet.
func NewPlayer(u User) *Player {
	return &Player{
		User:       u,
		ScoreSheet: map[Category]int{},
	}
}

// Game contains all data representing a game.
type Game struct {
	// Players has the list of the players in an ordered manner
	Players []*Player

	// Dices has the dices the game played with
	Dices []*Dice

	// Features has the features to play the game with
	Features []Feature

	// Round shows how many rounds were passed already.
	Round int

	// CurrentPlayer shows the index of the current player in the Players array.
	CurrentPlayer int

	// RollCount shows how many times the dices were rolled for the current user in this round.
	RollCount int

	Scorer *Score `json:"-"`

	Context map[string]interface{} `json:"-"`
}

// Feature represents the features available for the game.
type Feature string

// Available features
const (
	SixDice      Feature = "six-dice"
	TheChance    Feature = "the-chance"
	YahtzeeBonus Feature = "yahtzee-bonus"
	Equilizer    Feature = "equilizer"
)

func Features() []Feature {
	return []Feature{
		SixDice,
		YahtzeeBonus,
		TheChance,
		Equilizer,
	}
}

// NewGame initializes an empty Game.
func NewGame(features ...Feature) *Game {
	dices := NumberOfDices
	if features == nil {
		features = []Feature{}
	}
	if ContainsFeature(features, SixDice) {
		dices = 6
	}
	dd := make([]*Dice, dices)
	for i := 0; i < dices; i++ {
		dd[i] = &Dice{
			Value: 1,
		}
	}

	scorer := &Score{
		PreScoreActions: []func(game *Game){},
		ScoreActions:    NewDefaultScorer(),
		PostScoreActions: []func(game *Game){
			DefaultUpperSectionBonusAction,
		},
		PostGameActions: []func(game *Game){},
	}

	if ContainsFeature(features, TheChance) {
		scorer.PostGameActions = append(scorer.PostGameActions, TheChanceAction)
	}

	if ContainsFeature(features, YahtzeeBonus) {
		scorer.PreScoreActions = append(scorer.PreScoreActions, YahtzeeBonusPreScoreAction)
		scorer.PostScoreActions = append(scorer.PostScoreActions, YahtzeeBonusPostScoreAction)

		scorer.ScoreActions[FullHouse] = YahtzeeBonusFullHouse
		scorer.ScoreActions[SmallStraight] = YahtzeeBonusSmallStraight
		scorer.ScoreActions[LargeStraight] = YahtzeeBonusLargeStraight
	}

	if ContainsFeature(features, Equilizer) {
		scorer.PreScoreActions = append(scorer.PreScoreActions, EquilizerPreScoreAction)
		scorer.PostScoreActions = append(scorer.PostScoreActions, EquilizerPostScoreAction)
	}

	return &Game{
		Players:  []*Player{},
		Dices:    dd,
		Features: features,
		Scorer:   scorer,
		Context:  map[string]interface{}{},
	}
}

type User string

func NewUser(name string) *User {
	var u User
	u = User(name)
	return &u
}

func ContainsFeature(s []Feature, e Feature) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
