package structs

type TableStatus struct {
	StartingPhase   bool
	ActivePlayers   int
	RoundIndex      int
	Deck            Deck
	PlayerHandValue map[Player]int
	PlayerHands     map[Player]Hand
	Stand           map[Player]bool
	HasPlayed       map[Player]bool
	Winners         []string
}
