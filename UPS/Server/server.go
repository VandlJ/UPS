package main

import (
	_const "Server/const"
	"Server/structs"
	"Server/tools"
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pgaijin66/go-config-yaml/config"
)

var clientsMap = make(map[net.Conn]structs.Player)
var gameMap = make(map[string]structs.Game)
var playerPingMap = make(map[structs.Player]int)

var clientsMutex sync.Mutex
var gameMutex sync.Mutex

const ()

// main function: Entry point of the server application
// Initializes game rooms, starts pingPong handling in a separate goroutine,
// listens for incoming connections on a specified network and port.
// Handles accepted connections by spawning a goroutine for each connection.
func main() {
	gameMapInit()

	go pingPongInit()

	config, err := config.LoadConfiguration(_const.ConfigPath, _const.ConfigName, _const.ConfigType)
	if err != nil {
		log.Fatalf("Coudl not load configuration file: %v", err)
	}

	fmt.Println("Server config: ", config.Server)

	socket, err := net.Listen(_const.ConnType, _const.ConnHost+":"+_const.ConnPort)

	if err != nil {
		fmt.Println("LISTENING - ERR", err.Error())
		os.Exit(1)
	} else {
		fmt.Println("LISTENING - OK")
	}

	defer func(socket net.Listener) {
		socket.Close()

	}(socket)

	fmt.Println("Server is running...")

	for {
		client, err2 := socket.Accept()
		if err2 != nil {
			fmt.Println("ACCEPT - ERR", err2.Error())
			return
		}

		fmt.Println("Client " + client.RemoteAddr().String() + " connected.")

		go connHandler(client)
	}
}

// pingPongInit function: Starts a continuous loop for ping and pong communication
func pingPongInit() {
	for {
		ping()
		time.Sleep(_const.PingInterval * time.Second)
	}
}

// ping function: Sends ping messages to connected clients and manages disconnections based on ping limits
func ping() {
	clientsMutex.Lock()
	gameMutex.Lock()

	msg := tools.CreatePingMsg()

	clientsMapPingPong(msg)

	gameMapPingPong(msg)

	gameMutex.Unlock()
	clientsMutex.Unlock()
}

// clientsMapPingPong function: Manages ping-pong messages for clients in the map
// It iterates over clientsMap, increments ping count for each player, and sends a ping message.
// If the ping count exceeds the limit (_const.PingLimit), it closes the connection and removes the player from the map.
func clientsMapPingPong(msg string) {
	for conn, player := range clientsMap {
		playerPingMap[player]++
		tools.SendMsg(player.Socket, msg)
		if playerPingMap[player] > _const.PingLimit {
			conn.Close()
			delete(clientsMap, conn)
			delete(playerPingMap, player)
			continue
		}
	}
}

// gameMapPingPong function: Manages ping-pong messages for players in the game map
// It iterates over gameMap, managing player states based on ping count and sending ping messages.
// If the ping count exceeds the limit (_const.PingLimit), it closes the player's socket, removes the player,
// cancels messages, disconnects players, and initiates game phase restart in certain conditions.
func gameMapPingPong(msg string) {
	for gameID, game := range gameMap {
		for playerID, player := range game.Players {

			if _const.PingLowLimit < playerPingMap[player] && playerPingMap[player] <= _const.PingLimit {
				state := _const.Offline
				tools.PlayerStateSender(game, player, state)
				fmt.Println("Sending State Offline")
			} else if playerPingMap[player] == 0 {
				state := _const.Online
				tools.PlayerStateSender(game, player, state)
				fmt.Println("Sending State Online")
			}

			playerPingMap[player]++
			tools.SendMsg(player.Socket, msg)

			if playerPingMap[player] > _const.PingLimit {
				player.Socket.Close()
				delete(game.Players, playerID)
				delete(playerPingMap, player)
				cancelMsgSender(game)
				playerDisconnector(&game)

				game.GameData.StartingPhase = true
				gameMap[gameID] = game
				gameInfoBroadcaster()
			}
		}
	}
}

// cancelMsgSender function: Sends stop messages to all players in a game
// It creates a stop message and sends it to each player's socket in the game.
func cancelMsgSender(game structs.Game) {
	msg := tools.CreateStopMsg()
	for _, player := range game.Players {
		tools.SendMsg(player.Socket, msg)
	}
}

// gameMapInit function: Initializes game rooms and their initial states
// It populates gameMap with game rooms based on the defined count (_const.GameRoomsCount),
// initializing each room with an ID, an empty player map, and initial game data.
func gameMapInit() {
	gameMutex.Lock()
	for ID := 1; ID <= _const.GameRoomsCount; ID++ {
		gameID := fmt.Sprintf("Game %d", ID)
		gameMap[gameID] = structs.Game{
			ID:      gameID,
			Players: make(map[int]structs.Player),
			GameData: structs.TableStatus{
				StartingPhase:   true,
				PlayerHandValue: make(map[structs.Player]int),
				Stand:           make(map[structs.Player]bool),
				HasPlayed:       make(map[structs.Player]bool),
				ActivePlayers:   0,
				Winners:         make([]string, 0),
			},
		}
	}
	gameMutex.Unlock()
}

// connHandler function: Handles communication with a client after establishing a connection
// Reads incoming messages, validates their structure, and delegates message handling
func connHandler(client net.Conn) {
	reader := bufio.NewReader(client)

	for {
		readBuff, err := reader.ReadBytes('\n')

		if err != nil {
			clientsMutex.Lock()
			fmt.Println("Killing ", clientsMap[client].Nick)
			client.Close()
			clientsMutex.Unlock()
			fmt.Println("Client disconnected: ", client)
			return
		}

		msg := strings.TrimRight(string(readBuff), "\r\n")

		if msgValidator(msg) {
			fmt.Println("Msg OK")
			fmt.Println("Msg: ", msg)
			msgHandler(msg, client)
		} else {
			fmt.Println("Wrong msg - killing")
			client.Close()
			return
		}
	}
}

// msgHandler function: Handles different types of messages received from clients
// Parses messages, identifies commands, and takes appropriate actions
func msgHandler(msg string, client net.Conn) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	cmd := msg[len(_const.Pass)+_const.FormatLen : len(_const.Pass)+_const.FormatLen+_const.CmdLen]
	msgContent := msg[len(_const.Pass)+_const.FormatLen+_const.CmdLen:]

	switch cmd {
	case _const.Nick:
		if occupiedNick(msgContent, client) {
			playerReconnector(msgContent, client)
		} else if _, exists := clientsMap[client]; !exists && clientConn(client) == false {
			if nickCreator(client, msg) {
				fmt.Println("Client - nick ", clientsMap[client].Nick)
				gameInfoSender(client)
				return
			} else {
				fmt.Println("No nick - killing")
				client.Close()
			}
		} else {
			killer(client)
			return
		}
	case _const.Join, _const.Play, _const.GameTurn, _const.Pong:
		if _, exists := clientsMap[client]; exists || clientConn(client) {
			switch cmd {
			case _const.Join:
				playerJoiner(client, msg)
			case _const.Play:
				startGame(client)
			case _const.GameTurn:
				playerActionReceiver(client, msgContent)
			case _const.Pong:
				pingPong(client)
			}
			return
		} else {
			killer(client)
			return
		}
	default:
		fmt.Println("Killing - not recognised command: ", cmd)
		killer(client)
	}
}

// playerReconnector function: Handles reconnecting a player to a game after disconnection
// Restores player state in the game after reconnection
func playerReconnector(message string, client net.Conn) {
	game, player := getOccupiedNick(message, client)

	tempPlayerHands := game.GameData.PlayerHands[*player]
	delete(game.GameData.PlayerHands, *player)
	tempPlayerHandValue := game.GameData.PlayerHandValue[*player]
	delete(game.GameData.PlayerHandValue, *player)
	tempStand := game.GameData.Stand[*player]
	delete(game.GameData.Stand, *player)
	tempHasPlayed := game.GameData.HasPlayed[*player]
	delete(game.GameData.HasPlayed, *player)
	tempPlayerPingMap := playerPingMap[*player]
	delete(playerPingMap, *player)

	if player != nil && game != nil {
		tools.KillerMsgSender(player.Socket)
		player.Socket.Close()

		player.Socket = client
		if game.GameData.StartingPhase == false {
			game.GameData.PlayerHands[*player] = tempPlayerHands
			game.GameData.PlayerHandValue[*player] = tempPlayerHandValue
			game.GameData.Stand[*player] = tempStand
			game.GameData.HasPlayed[*player] = tempHasPlayed
			playerPingMap[*player] = tempPlayerPingMap

			playerPingMap[*player] = 0
		}

		for idx, p := range game.Players {
			if p.Nick == player.Nick {
				game.Players[idx] = *player
				break
			}
		}

		gameMap[game.ID] = *game

		msg := tools.CreateJoinMsg(true)
		tools.SendMsg(player.Socket, msg)

		startCheckSender(*player, *game)

		if !gameMap[game.ID].GameData.StartingPhase {
			clientInfoResender(client)
		}
	} else {
		fmt.Println("Player or game not found")
	}
}

// startCheckSender function: Sends information about the game start to a reconnected player
func startCheckSender(player structs.Player, game structs.Game) {
	msg := tools.CreateCheckMsg(gameStartChecker(game), len(game.Players), _const.MaxPlayers)
	tools.SendMsg(player.Socket, msg)
}

// clientInfoResender function: Resends game information to a reconnected client after disconnection
func clientInfoResender(client net.Conn) {
	player := getClientConn(client)
	lobbyID := playerGameFinder(*player).ID
	lobby, _ := gameMap[lobbyID]
	messageFinal := tools.CreateReconnectMsg(&lobby, *player)
	tools.SendMsg(client, messageFinal)
}

// killer function: Sends a message and closes the client connection
func killer(client net.Conn) {
	msg := "Killing"
	tools.SendMsg(client, msg)
	client.Close()
}

// pingPong function: Resets the ping count for a client upon receiving a pong message
func pingPong(client net.Conn) {
	for _, game := range gameMap {
		for _, player := range game.Players {
			if player.Socket == client {
				playerPingMap[player] = 0
				return
			}
		}
	}

	for _, player := range clientsMap {
		if player.Socket == client {
			playerPingMap[player] = 0
			return
		}
	}
}

// playerActionReceiver function: Handles player actions received from clients during gameplay
// Processes player actions like "HIT" or "STAND" and updates game state accordingly
func playerActionReceiver(client net.Conn, msg string) {
	player := getClientConn(client)

	if msg != "STAND" && msg != "HIT" {
		fmt.Println("Invalid action")
		killer(client)
		return
	}

	if player == nil {
		fmt.Println("Player not found")
		killer(client)
		return
	}

	gameID := playerGameFinder(*player).ID
	game, ok := gameMap[gameID]
	if ok {
		gameMutex.Lock()
		playerActionHandler(&game, *player, msg, gameID)
		if game.GameData.StartingPhase {
			fmt.Println("Game has ended")
			gameInfoBroadcaster()
		}
		gameMutex.Unlock()
	}
}

// playerActionHandler function: Handles player actions during the game
// Updates game state based on the actions taken by players (hitting or standing)
func playerActionHandler(game *structs.Game, player structs.Player, turn string, gameID string) {

	if turn == "HIT" && !game.GameData.HasPlayed[player] {
		game.GameData.HasPlayed[player] = true
		if !game.GameData.Stand[player] {
			deckSize := len(game.GameData.Deck.Cards)
			fmt.Printf("HIT - Deck size: %d\n", deckSize)

			newCard := cardDealer(&game.GameData.Deck, 1)
			fmt.Println("New cards:", newCard)

			existingHand := structs.Hand{
				Cards: make([]structs.Card, len(game.GameData.PlayerHands[player].Cards)),
			}
			copy(existingHand.Cards, game.GameData.PlayerHands[player].Cards)

			existingHand.Cards = append(existingHand.Cards, newCard.Cards...)

			game.GameData.PlayerHands[player] = existingHand

			handValueCalculator(&game.GameData, player)

			game.GameData.RoundIndex += 1

			gameMap[gameID] = *game

			for _, player2 := range gameMap[gameID].Players {
				msgToBroadcast := tools.CreateTurnMsg(*game, player2)
				fmt.Print("Message to broadcast: ", msgToBroadcast)
				tools.SendMsg(player2.Socket, msgToBroadcast)
			}

			if game.GameData.RoundIndex%game.GameData.ActivePlayers == 0 {
				for _, player2 := range gameMap[gameID].Players {
					game.GameData.HasPlayed[player2] = false
					gameMap[gameID] = *game
					msgToBroadcast := tools.CreateNextMsg(*game, player2)
					fmt.Print("Message to broadcast: ", msgToBroadcast)
					tools.SendMsg(player2.Socket, msgToBroadcast)
				}
			}
		} else {
			for _, player2 := range gameMap[gameID].Players {
				msgToBroadcast := tools.CreateTurnMsg(*game, player2)
				fmt.Print("Message to broadcast: ", msgToBroadcast)
				tools.SendMsg(player2.Socket, msgToBroadcast)

				game.GameData.RoundIndex += 1
				gameMap[gameID] = *game
			}
		}
	} else if turn == "STAND" && !game.GameData.HasPlayed[player] {
		game.GameData.HasPlayed[player] = true
		game.GameData.Stand[player] = true
		if game.GameData.Stand[player] {
			fmt.Println("Stand status: ", game.GameData.Stand)

			game.GameData.ActivePlayers -= 1
			game.GameData.HasPlayed[player] = true
			gameMap[gameID] = *game

			if game.GameData.ActivePlayers == 0 {
				fmt.Println("Game is ending")

				winner := whoIsTheWinner(game.GameData)
				if winner != nil {
					game.GameData.Winners = append(game.GameData.Winners, winner.Nick)
				}

				for _, player2 := range gameMap[gameID].Players {
					msgToBroadcast := tools.CreateEndMsg(*game)
					fmt.Print("Message to broadcast: ", msgToBroadcast)
					tools.SendMsg(player2.Socket, msgToBroadcast)
				}
				game.GameData.Winners = nil
				playerDisconnector(game)
				game.GameData.StartingPhase = true
				gameMap[gameID] = *game

			} else {
				game.GameData.RoundIndex += 1

				if game.GameData.RoundIndex%game.GameData.ActivePlayers == 0 {
					for _, player2 := range gameMap[gameID].Players {
						game.GameData.HasPlayed[player2] = false
						gameMap[gameID] = *game
						msgToBroadcast := tools.CreateNextMsg(*game, player2)
						fmt.Print("Message to broadcast: ", msgToBroadcast)
						tools.SendMsg(player2.Socket, msgToBroadcast)
					}
				}
			}
		}
		game.GameData.Stand[player] = true
		game.GameData.HasPlayed[player] = true
		gameMap[gameID] = *game
	} else {
		fmt.Println("Player action handler error")
		killer(player.Socket)
		return
	}
}

// playerDisconnector function: Disconnects players from a game and resets game data
func playerDisconnector(game *structs.Game) {
	for _, player := range game.Players {
		clientsMap[player.Socket] = player
	}
	game.Players = make(map[int]structs.Player)
}

// whoIsTheWinner function: Determines the winner based on player hand values
func whoIsTheWinner(gameData structs.TableStatus) *structs.Player {
	var winner *structs.Player
	highestScore := 0

	for player, score := range gameData.PlayerHandValue {
		if score > highestScore && score <= 21 {
			highestScore = score
			playerCopy := player
			winner = &playerCopy
		}
		if winner == nil {
			// no winner
		}
	}
	return winner
}

// startGame function: Initiates the start of the game based on player count
func startGame(client net.Conn) {
	player := getClientConn(client)
	if player == nil {
		fmt.Println("Player not found.")
		killer(client)
		return
	}

	game := playerGameFinder(*player)
	if game == nil {
		fmt.Println("Game not found.")
		return
	} else if gameStartChecker(*game) {
		gameStartHandler(game.ID)
		gameInfoBroadcaster()
	} else {
		fmt.Println("Could not switch to game - not enough players.")
		killer(client)
		return
	}
}

// gameStartHandler function: Handles the start of a game, deals cards, and initiates gameplay
func gameStartHandler(gameID string) {
	fmt.Println("Starting game ", gameID)
	gameMutex.Lock()
	defer gameMutex.Unlock()

	if existingGame, ok := gameMap[gameID]; ok {
		existingGame.GameData.StartingPhase = false

		existingGame.GameData.Deck = createDeck()
		fmt.Println("Deck created")

		existingGame.GameData.Deck = shuffleDeck(existingGame.GameData.Deck)
		fmt.Println("Deck shuffled")

		existingGame.GameData.PlayerHands = make(map[structs.Player]structs.Hand)
		fmt.Println("Player hands ready")

		for _, player := range existingGame.Players {
			existingGame.GameData.Stand[player] = false
			existingGame.GameData.HasPlayed[player] = false
			initialHand := cardDealer(&existingGame.GameData.Deck, 2)
			existingGame.GameData.PlayerHands[player] = initialHand
			handValueCalculator(&existingGame.GameData, player)
		}

		playerHandsPrinter(existingGame.GameData.PlayerHands)

		existingGame.GameData.ActivePlayers = len(existingGame.Players)
		gameMap[gameID] = existingGame

		for _, player := range gameMap[gameID].Players {
			msgToBroadcast := tools.CreateInitMsg(existingGame, player)
			tools.SendMsg(player.Socket, msgToBroadcast)
			existingGame.GameData.RoundIndex = 0
		}
		return
	}
}

// handValueCalculator function: Calculates the total value of a player's hand
func handValueCalculator(gameData *structs.TableStatus, player structs.Player) {
	hand := gameData.PlayerHands[player]
	totalValue := 0

	for _, card := range hand.Cards {
		totalValue += card.Value
	}
	gameData.PlayerHandValue[player] = totalValue
}

// cardDealer function: Deals cards to players from the deck
func cardDealer(deck *structs.Deck, cardsCount int) structs.Hand {
	var hand structs.Hand

	fmt.Printf("Deck size: %d\nCards dealt: %d\n", len(deck.Cards), cardsCount)

	if len(deck.Cards) < cardsCount {
		fmt.Println("Not enough cards in deck")
		return hand
	}

	hand.Cards = deck.Cards[:cardsCount]
	*deck = structs.Deck{Cards: deck.Cards[cardsCount:]}

	fmt.Printf("Deck size: %d\n", len(deck.Cards))

	return hand
}

// shuffleDeck function: Shuffles the deck of cards
func shuffleDeck(deck structs.Deck) structs.Deck {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))

	numOfCards := len(deck.Cards)
	shuffledDeck := make([]structs.Card, numOfCards)
	perm := random.Perm(numOfCards)

	for i, j := range perm {
		shuffledDeck[j] = deck.Cards[i]
	}

	return structs.Deck{Cards: shuffledDeck}
}

// createDeck function: Creates a standard deck of cards
func createDeck() structs.Deck {
	var deck structs.Deck

	suits := []string{"Hearts", "Diamonds", "Clubs", "Spades"}
	values := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 10, 10, 10}

	for _, suit := range suits {
		for i := 0; i < len(values); i++ {
			card := structs.Card{Suit: suit, Value: values[i]}
			deck.Cards = append(deck.Cards, card)
		}
	}
	return deck
}

// playerHandsPrinter function: Prints the hands of players during gameplay
func playerHandsPrinter(playerHands map[structs.Player]structs.Hand) {
	for player, hand := range playerHands {
		fmt.Printf("Player %v has cards: %v\n", player, hand.Cards)
	}
}

// playerGameFinder function: Finds the game a player is currently associated with
func playerGameFinder(player structs.Player) *structs.Game {
	gameMutex.Lock()
	defer gameMutex.Unlock()
	for _, game := range gameMap {
		for _, p := range game.Players {
			if p == player {
				return &game
			}
		}
	}
	return nil
}

// clientConn function: Checks if a client is connected to any game
func clientConn(client net.Conn) bool {
	gameMutex.Lock()
	defer gameMutex.Unlock()
	for _, gameState := range gameMap {
		for _, player := range gameState.Players {
			if player.Socket == client {
				return true
			}
		}
	}
	return false
}

// getClientConn function: Gets the player associated with a client connection
func getClientConn(client net.Conn) *structs.Player {
	gameMutex.Lock()
	defer gameMutex.Unlock()
	for _, gameState := range gameMap {
		for _, player := range gameState.Players {
			if player.Socket == client {
				return &player
			}
		}
	}
	return nil
}

// getOccupiedNick function: Checks if a nickname is occupied by another player in a different game
func getOccupiedNick(nick string, socket net.Conn) (*structs.Game, *structs.Player) {
	for _, game := range gameMap {
		for _, player := range game.Players {
			if player.Nick == nick && player.Socket != socket {
				return &game, &player
			}
		}
	}
	return nil, nil
}

// occupiedNick function: Checks if a nickname is occupied by another player in the same game
func occupiedNick(nick string, socket net.Conn) bool {
	for _, game := range gameMap {
		for _, player := range game.Players {
			if player.Nick == nick {
				return player.Socket != socket
			}
		}
	}
	return false
}

// playerJoiner function: Handles player joining a game lobby
func playerJoiner(client net.Conn, msg string) {
	gameName := msg[len(_const.Pass)+_const.FormatLen+_const.CmdLen:]
	gameMutex.Lock()
	if game, ok := gameMap[gameName]; ok {
		if isGameNotFull(game) {
			tryJoin(game, client, gameName)
		} else {
			fmt.Println("Lobby is not empty.")
		}
	} else {
		fmt.Printf("Lobby %s not found in game map.\n", gameName)
		killer(client)
	}
	gameMutex.Unlock()
}

// tryJoin function: Tries to join a player to a game if the lobby is not full
func tryJoin(game structs.Game, client net.Conn, gameName string) {
	if _, exists := clientsMap[client]; exists {
		playerID := len(game.Players) + 1
		game.Players[playerID] = clientsMap[client]
		fmt.Printf("User %s has joined the game %s\n", clientsMap[client].Nick, gameName)
		delete(clientsMap, client)
		playerMover(game.Players[playerID])
		gameInfoBroadcaster()
		startInfoSender(game)
	} else {
		fmt.Println("User not found in clients map.")
		killer(client)
		return
	}
}

// startInfoSender function: Sends game start information to players in a game lobby
func startInfoSender(game structs.Game) {
	for _, player := range game.Players {
		gameMutex.Unlock()
		msg := tools.CreateCheckMsg(gameStartChecker(game), len(game.Players), _const.MaxPlayers)
		tools.SendMsg(player.Socket, msg)
		gameMutex.Lock()
	}
}

// gameStartChecker function: Checks if the game is ready to start based on player count
func gameStartChecker(game structs.Game) bool {
	gameMutex.Lock()
	defer gameMutex.Unlock()
	return len(game.Players) >= 2 && game.GameData.StartingPhase
}

// gameInfoBroadcaster function: Broadcasts game information to all connected clients
func gameInfoBroadcaster() {
	for _, player := range clientsMap {
		gameMutex.Unlock()
		gameInfoSender(player.Socket)
		gameMutex.Lock()
	}
}

// playerMover function: Sends a join message to the specified player
// It creates a join message and sends it to the player's socket using tools.SendMsg
func playerMover(player structs.Player) {
	msg := tools.CreateJoinMsg(true)
	tools.SendMsg(player.Socket, msg)
}

// isGameNotFull function: Checks if the number of players in the game is less than the maximum allowed players
// Returns true if the number of players in the game is less than the maximum limit defined by _const.MaxPlayers
func isGameNotFull(game structs.Game) bool {
	return len(game.Players) < _const.MaxPlayers
}

// gameInfoSender function: Sends game information to a client
func gameInfoSender(client net.Conn) {
	pass := _const.Pass
	cmd := _const.GamesInfo

	gameMutex.Lock()
	var gameStrings []string
	for _, game := range gameMap {
		playerCount := len(game.Players)
		StartingPhase := 0
		if game.GameData.StartingPhase {
			StartingPhase = 1
		}
		gameString := fmt.Sprintf("%s|%d|%d|%d", game.ID, _const.MaxPlayers, playerCount, StartingPhase)
		gameStrings = append(gameStrings, gameString)
	}

	gameMutex.Unlock()
	msgCnt := strings.Join(gameStrings, ";")
	messageLength := fmt.Sprintf("%03d", len(msgCnt))
	msg := pass + messageLength + cmd + msgCnt + "\n"
	fmt.Print("Sending: ", msg)
	gameMutex.Lock()
	tools.SendMsg(client, msg)
	gameMutex.Unlock()
}

// nickCreator function: Creates a nickname for a client based on the received message
// If the message type is a nickname command, it assigns a nickname to the client
// Checks for existing nicknames and deletes the corresponding client if the nickname is already in use
func nickCreator(client net.Conn, msg string) bool {
	cmd := msg[len(_const.Pass)+_const.FormatLen : len(_const.Pass)+_const.FormatLen+_const.CmdLen]
	if cmd == _const.Nick {
		nick := msg[len(_const.Pass)+_const.FormatLen+_const.CmdLen:]
		for oldClient, oldPlayer := range clientsMap {
			if oldPlayer.Nick == nick {
				tools.KillerMsgSender2(oldPlayer.Socket)
				delete(clientsMap, oldClient)
				break
			}
		}

		clientsMap[client] = structs.Player{
			Nick:   msg[len(_const.Pass)+_const.FormatLen+_const.CmdLen:],
			Socket: client,
		}
		return true
	} else {
		return false
	}
}

// msgValidator function: Validates the incoming message format and content integrity
// Checks if the received message has the correct password, format, and length
func msgValidator(msg string) bool {
	if len(msg) < (len(_const.Pass) + _const.CmdLen + _const.FormatLen) {
		return false
	}

	pass := msg[:len(_const.Pass)]

	if pass != _const.Pass {
		fmt.Printf("Received password: %s, System password: %s\n", pass, _const.Pass)
		return false
	}

	stringLen := msg[len(_const.Pass) : len(_const.Pass)+_const.FormatLen]
	length, err := strconv.Atoi(stringLen)
	if err != nil {
		return false
	}
	if length != len(msg)-len(_const.Pass)-_const.FormatLen-_const.CmdLen {
		fmt.Printf("Length from message: %d, calculated length: %d\n", length, len(msg)-len(_const.Pass)-_const.FormatLen-_const.CmdLen)
		return false
	}
	return true
}
