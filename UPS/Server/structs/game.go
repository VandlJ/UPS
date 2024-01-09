package structs

// Game represents the structure defining a game.
// It contains an ID identifying the game uniquely,
// a map of Players, where the key is an integer representing the player ID,
// and the value is a Player struct representing a player in the game,
// and GameData, which stores the TableStatus struct representing the game status.
type Game struct {
	ID       string         // Unique identifier for the game
	Players  map[int]Player // Map of player IDs to Player structs representing players in the game
	GameData TableStatus    // TableStatus struct representing the game status
}
