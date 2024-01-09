package _const

const (
	// InvalidExecutingParameters represents the error message when invalid parameter count is included
	InvalidExecutingParameters = "Run it with 2 other parameters"

	// ConfigPath represents the config path name
	// ConfigPath = "../Server/data"

	// ConfigName represents the config file name
	// ConfigName = "config"

	// ConfigType represents the config file type
	// ConfigType = "yaml"

	// ConnType represents the connection type used (TCP in this case)
	ConnType = "tcp"

	// ConnHost represents the host address (defaulting to localhost)
	// ConnHost = "0.0.0.0" // 172.24.32.1 // 147.228.67.103

	// ConnPort represents the port number used for the connection
	// ConnPort = "10000"

	// GameRoomsCount represents the total number of game rooms available
	GameRoomsCount = 3

	// MaxPlayers represents the maximum number of players allowed in a game room
	MaxPlayers = 8

	// PingInterval represents the interval (in seconds) for sending ping messages
	PingInterval = 500

	// PingLimit represents the maximum number of allowed pings without a response
	PingLimit = 10000

	// PingLowLimit represents the lower limit of pings
	PingLowLimit = 1000
)
