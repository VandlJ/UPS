package structs

type Game struct {
	ID       string
	Players  map[int]Player
	GameData TableStatus
}
