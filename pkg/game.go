package yahtzee

// Box represents one field on the scoring sheet.
type Box struct {
	// Filled shows whether the Box has a value already or it's still
	// available to choose.
	Filled bool

	// Score keeps the value of the box if it is filled.
	Score int
}

// Sheet represents a standard scoring sheet.
type Sheet struct {
	Ones, Twos, Threes, Fours, Fives, Sixes Box
	ThreeOfAKind, FourOfAKind               Box
	FullHouse                               Box
	SmallStraight, LargeStraight            Box
	Yahtzee                                 Box
	Chance                                  Box
}

// Player contains all data representing a player.
type Player struct {
	Name   string
	Scores *Sheet
}

// Game contains all data representing a game.
type Game struct {
	Players []*Player

	// Round shows how many rounds were passed already.
	Round int

	// Current shows the index of the current player in the Players array.
	Current int
}
