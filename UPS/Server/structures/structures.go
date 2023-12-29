package structures

import "net"

type Game struct {
	ID       string
	Players  map[int]Player
	GameData GameState
}

type Player struct {
	Socket   net.Conn
	Nickname string
}

type GameState struct {
	IsLobby         bool
	PlayerHandValue map[Player]int
	PlayerHands     map[Player]Hand
	RoundIndex      int
	Deck            Deck
	Stand           bool
}

type Card struct {
	Rank  string
	Suit  string
	Value int
}

type Deck struct {
	Cards []Card
}

type Hand struct {
	Cards []Card
}
