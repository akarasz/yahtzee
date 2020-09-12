package models

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
	Twos            = "twos"
	Threes          = "threes"
	Fours           = "fours"
	Fives           = "fives"
	Sixes           = "sixes"
	Bonus           = "bonus"

	ThreeOfAKind  = "three-of-a-kind"
	FourOfAKind   = "four-of-a-kind"
	FullHouse     = "full-house"
	SmallStraight = "small-straight"
	LargeStraight = "large-straight"
	Yahtzee       = "yahtzee"
	Chance        = "chance"
)

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

	// Round shows how many rounds were passed already.
	Round int

	// CurrentPlayer shows the index of the current player in the Players array.
	CurrentPlayer int

	// RollCount shows how many times the dices were rolled for the current user in this round.
	RollCount int
}

// NewGame initializes an empty Game.
func NewGame() *Game {
	dd := make([]*Dice, NumberOfDices)
	for i := 0; i < NumberOfDices; i++ {
		dd[i] = &Dice{
			Value: 1,
		}
	}

	return &Game{
		Players: []*Player{},
		Dices:   dd,
	}
}

type User string

func NewUser(name string) *User {
	var u User
	u = User(name)
	return &u
}
