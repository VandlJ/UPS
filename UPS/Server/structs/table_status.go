package structs

// TableStatus represents the status of the game table.
// It includes various attributes such as the game phase, active players,
// round index, deck of cards, player hand values, player hands, stand status of players,
// whether players have played, and a list of winners.
type TableStatus struct {
	StartingPhase   bool            // Indicates if the game is in the starting phase
	ActivePlayers   int             // Number of active players
	RoundIndex      int             // Index of the current round
	Deck            Deck            // Deck of cards in the game
	PlayerHandValue map[Player]int  // Map of player hand values
	PlayerHands     map[Player]Hand // Map of player hands
	Stand           map[Player]bool // Map indicating if players have stood
	HasPlayed       map[Player]bool // Map indicating if players have played
	Winners         []string        // List of winners
}
