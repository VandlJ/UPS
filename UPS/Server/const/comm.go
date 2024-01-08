package _const

const (
	// Pass represents a password or authentication token
	Pass = "420BJ69"

	// FormatLen represents the length of the format section in a message
	FormatLen = 3

	// CmdLen represents the length of the command section in a message
	CmdLen = 4

	// Nick represents a command for setting a nickname
	Nick = "NICK"

	// Ping represents a ping command
	Ping = "PING"

	// Pong represents a pong command
	Pong = "PONG"

	// Join represents a join command
	Join = "JOIN"

	// Play represents a play command
	Play = "PLAY"

	// GamesInfo represents a command for obtaining game information
	GamesInfo = "GMIF"

	// GameJoin represents a command for joining a game
	GameJoin = "GMJN"

	// GameStartCheck represents a command for checking if a game can start
	GameStartCheck = "GMCK"

	// GameStart represents a command for starting a game
	GameStart = "GMST"

	// GameEnd represents a command for ending a game
	GameEnd = "GMEN"

	// GameTurn represents a command for taking a turn in a game
	GameTurn = "TURN"

	// GameNextRound represents a command for initiating the next round in a game
	GameNextRound = "NEXT"

	// Stop represents a command to stop an action or process
	Stop = "STOP"

	// RetrieveState represents a command for retrieving the state of a game
	RetrieveState = "RETR"

	// State represents a command for setting the state
	State = "STAT"

	// Kill represents a command for terminating or stopping something
	Kill = "KILL"

	// Kill2 represents another command for terminating or stopping something (variation)
	Kill2 = "KIL2"

	// Offline represents the status indicator for being offline
	Offline = "0"

	// Online represents the status indicator for being online
	Online = "1"
)
