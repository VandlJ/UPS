package _const

const (
	// ConnType represents the connection type used (TCP in this case)
	ConnType = "tcp"

	// ConnHost represents the host address (defaulting to localhost)
	ConnHost = "localhost" // 172.24.32.1 // 147.228.67.103

	// ConnPort represents the port number used for the connection
	ConnPort = "10000"

	// GameRoomsCount represents the total number of game rooms available
	GameRoomsCount = 3

	// MaxPlayers represents the maximum number of players allowed in a game room
	MaxPlayers = 8

	// PingInterval represents the interval (in seconds) for sending ping messages
	PingInterval = 5

	// PingLimit represents the maximum number of allowed pings without a response
	PingLimit = 5

	// PingLowLimit represents the lower limit of pings
	PingLowLimit = 0
)
