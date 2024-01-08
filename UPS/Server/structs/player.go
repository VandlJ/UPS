package structs

import "net"

// Player represents a player in the game.
// It contains a socket connection and a nickname.
type Player struct {
	Socket net.Conn // Socket connection associated with the player
	Nick   string   // Nickname of the player
}
