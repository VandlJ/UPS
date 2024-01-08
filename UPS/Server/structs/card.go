package structs

// Card represents a structure defining a playing card in a game.
// It contains fields for the rank, suit, and numerical value of the card.
type Card struct {
	Rank  string // Rank of the card (e.g., "Ace", "King", "Queen", "Jack")
	Suit  string // Suit of the card (e.g., "Hearts", "Diamonds", "Clubs", "Spades")
	Value int    // Numerical value of the card (e.g., 2-10 for numbered cards, 11-13 for face cards)
}
