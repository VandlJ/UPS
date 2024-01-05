package main

import (
	"Server/const"
	"Server/structs"
	"Server/tools"
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var clientsMap = make(map[net.Conn]structs.Player)
var gameMap = make(map[string]structs.Game)
var playerPingMap = make(map[structs.Player]int)

var clientsMutex sync.Mutex
var gameMutex sync.Mutex

func main() {
	gameMapInit()

	go pingPongInit()

	socket, err := net.Listen(_const.ConnType, _const.ConnHost+":"+_const.ConnPort)

	if err != nil {
		fmt.Println("LISTENING - ERR", err.Error())
		os.Exit(1)
	} else {
		fmt.Println("LISTENING - OK")
	}
	defer func(socket net.Listener) {
		err := socket.Close()
		if err != nil {
			return
		}
	}(socket)

	fmt.Println("Server is running...")

	for {
		client, err := socket.Accept()
		if err != nil {
			fmt.Println("ACCEPT - ERR", err.Error())
			return
		}

		fmt.Println("Client " + client.RemoteAddr().String() + " connected.")

		// ping pong interval setter
		pingPongIntervalSetter(client)

		go connHandler(client)
	}
}

func pingPongIntervalSetter(client net.Conn) {
	msgToBroadcast := tools.PingPongIntervalMsg()
	fmt.Println("Message to broadcast: ", msgToBroadcast)
	tools.SendMsg(client, msgToBroadcast)
}

func pingPongInit() {
	for {
		fmt.Println("Sending PING")
		ping()
		time.Sleep(_const.PingInterval * time.Second)
	}
}

func ping() {
	clientsMutex.Lock()
	gameMutex.Lock()

	msg := tools.CreatePingMessage()

	clientsMapPingPong(msg)

	gameMapPingPong(msg)

	gameMutex.Unlock()
	clientsMutex.Unlock()
}

func clientsMapPingPong(msg string) {
	for conn, player := range clientsMap {
		playerPingMap[player]++
		tools.SendMsg(player.Socket, msg)
		if playerPingMap[player] > 10 {
			conn.Close()
			delete(clientsMap, conn)
			delete(playerPingMap, player)
			continue
		}
	}
}

func gameMapPingPong(msg string) {
	for gameID, game := range gameMap {
		for playerID, player := range game.Players {
			playerPingMap[player]++
			tools.SendMsg(player.Socket, msg)
			if playerPingMap[player] > 10 {
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

func cancelMsgSender(game structs.Game) {
	msg := tools.CreateCancelMessage()
	for _, player := range game.Players {
		tools.SendMsg(player.Socket, msg)
	}
}

func gameMapInit() {
	gameMutex.Lock()
	for ID := 1; ID <= _const.GameRoomsCount; ID++ {
		gameID := fmt.Sprintf("game%d", ID)
		gameMap[gameID] = structs.Game{
			ID:      gameID,
			Players: make(map[int]structs.Player),
			GameData: structs.TableStatus{
				StartingPhase:   true,
				PlayerHandValue: make(map[structs.Player]int),
				Stand:           make(map[structs.Player]bool),
				ActivePlayers:   0,
				Winners:         make([]string, 0),
			},
		}
	}
	gameMutex.Unlock()
}

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
			fmt.Println("Message structure is valid.")
			fmt.Println("Message: ", msg)
			msgHandler(msg, client)
		} else {
			fmt.Println("Message structure is invalid. Closing connection.")
			return
		}
	}
}

func msgHandler(msg string, client net.Conn) {
	clientsMutex.Lock()
	if _, exists := clientsMap[client]; !exists && clientConn(client) == false {
		if nickCreator(client, msg) {
			fmt.Println("Client connected as", clientsMap[client].Nick)
			gameInfoSender(client)
		} else {
			fmt.Println("No nick has been set")
			err := client.Close()
			if err != nil {
				return
			}
		}
	} else {
		cmd := msg[len(_const.Pass)+_const.FormatLen : len(_const.Pass)+_const.FormatLen+_const.CmdLength]
		msgContent := msg[len(_const.Pass)+_const.FormatLen+_const.CmdLength:]

		switch cmd {
		case "JOIN":
			playerJoiner(client, msg)
		case "PLAY":
			startGame(client)
		case "TURN":
			playerActionReceiver(client, msgContent)
		case "PONG":
			pingPong(client)
		default:
			fmt.Println("Unknown command: ", cmd)
			killer(client)
		}
	}
	clientsMutex.Unlock()
}

func killer(client net.Conn) {
	msg := "Killing"
	tools.SendMsg(client, msg)
	client.Close()
}

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

func playerActionReceiver(client net.Conn, msg string) {
	player := clientConnReturn(client)

	if msg != "STAND" && msg != "HIT" {
		fmt.Println("Invalid action")
		return
	}

	if player == nil {
		fmt.Println("Could not find specified player")
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

func playerActionHandler(game *structs.Game, player structs.Player, turn string, gameID string) {
	fmt.Printf("%s has played %s.\n", player.Nick, turn)
	fmt.Println("Active players: ", game.GameData.ActivePlayers)

	if turn == "HIT" {
		if game.GameData.Stand[player] == false {
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

			fmt.Println("HIT - Hand: ", game.GameData.PlayerHands[player])

			handValueCalculator(&game.GameData, player)

			game.GameData.RoundIndex += 1

			gameMap[gameID] = *game

			for _, player := range gameMap[gameID].Players {
				msgToBroadcast := tools.PlayerActionMsg(*game, player)
				fmt.Println("Message to broadcast: ", msgToBroadcast)
				tools.SendMsg(player.Socket, msgToBroadcast)
			}

			fmt.Println("RoundIndex: ", game.GameData.RoundIndex)
			fmt.Println("Active PlayerCount: ", game.GameData.ActivePlayers)

			if game.GameData.RoundIndex%game.GameData.ActivePlayers == 0 {
				fmt.Println("Every player has played")
				for _, player := range gameMap[gameID].Players {
					msgToBroadcast := tools.NextRoundMsg(*game, player)
					fmt.Println("Message to broadcast: ", msgToBroadcast)
					tools.SendMsg(player.Socket, msgToBroadcast)
				}
			}
		} else {
			for _, player := range gameMap[gameID].Players {
				msgToBroadcast := tools.PlayerActionMsg(*game, player)
				fmt.Println("Message to broadcast: ", msgToBroadcast)
				tools.SendMsg(player.Socket, msgToBroadcast)

				game.GameData.RoundIndex += 1
				gameMap[gameID] = *game
			}
		}
	} else if turn == "STAND" {
		game.GameData.Stand[player] = true
		if game.GameData.Stand[player] == true {
			fmt.Println("Stand status: ", game.GameData.Stand)

			game.GameData.ActivePlayers -= 1
			gameMap[gameID] = *game

			if game.GameData.ActivePlayers == 0 {
				fmt.Println("Game is ending")

				winner := whoIsTheWinner(game.GameData)
				if winner != nil {
					fmt.Println("Winner is: ", winner)
					game.GameData.Winners = append(game.GameData.Winners, winner.Nick)
				}

				for _, player := range gameMap[gameID].Players {
					msgToBroadcast := tools.EndMsg(*game)
					fmt.Println("Message to broadcast: ", msgToBroadcast)
					tools.SendMsg(player.Socket, msgToBroadcast)
				}

				playerDisconnector(game)
				game.GameData.StartingPhase = true
				gameMap[gameID] = *game

			} else {
				fmt.Println("RoundIndex: ", game.GameData.RoundIndex)
				fmt.Println("Active PlayerCount: ", game.GameData.ActivePlayers)

				if game.GameData.RoundIndex%game.GameData.ActivePlayers == 0 {
					fmt.Println("Every player has played")
					for _, player := range gameMap[gameID].Players {
						msgToBroadcast := tools.NextRoundMsg(*game, player)
						fmt.Println("Message to broadcast: ", msgToBroadcast)
						tools.SendMsg(player.Socket, msgToBroadcast)
					}
				}
			}
		}
	} else {
		fmt.Println("Player action handler error")
		return
	}
}

func playerDisconnector(game *structs.Game) {
	for _, player := range game.Players {
		clientsMap[player.Socket] = player
	}
	game.Players = make(map[int]structs.Player)
}

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
	fmt.Println("Winner: ", winner)
	return winner
}

func startGame(client net.Conn) {
	player := clientConnReturn(client)
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
			initialHand := cardDealer(&existingGame.GameData.Deck, 2)
			existingGame.GameData.PlayerHands[player] = initialHand
			handValueCalculator(&existingGame.GameData, player)
		}

		playerHandsPrinter(existingGame.GameData.PlayerHands)

		existingGame.GameData.ActivePlayers = len(existingGame.Players)
		gameMap[gameID] = existingGame

		for _, player := range gameMap[gameID].Players {
			msgToBroadcast := tools.InitMsg(existingGame, player)
			tools.SendMsg(player.Socket, msgToBroadcast)
			existingGame.GameData.RoundIndex = 0
		}
		return
	}
}

func handValueCalculator(gameData *structs.TableStatus, player structs.Player) {
	hand := gameData.PlayerHands[player]
	totalValue := 0

	for _, card := range hand.Cards {
		totalValue += card.Value
	}
	gameData.PlayerHandValue[player] = totalValue
}

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

func createDeck() structs.Deck {
	var deck structs.Deck

	suits := []string{"Hearts", "Diamonds", "Clubs", "Spades"}
	values := []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 10, 10, 10, 11}

	for _, suit := range suits {
		for i := 0; i < len(values); i++ {
			card := structs.Card{Suit: suit, Value: values[i]}
			deck.Cards = append(deck.Cards, card)
		}
	}
	return deck
}

func playerHandsPrinter(playerHands map[structs.Player]structs.Hand) {
	for player, hand := range playerHands {
		fmt.Printf("Player %v has cards: %v\n", player, hand.Cards)
	}
}

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

func clientConnReturn(client net.Conn) *structs.Player {
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

func playerJoiner(client net.Conn, msg string) {
	gameName := msg[len(_const.Pass)+_const.FormatLen+_const.CmdLength:]
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
	}
}

func startInfoSender(game structs.Game) {
	for _, player := range game.Players {
		gameMutex.Unlock()
		msg := tools.GameReady(gameStartChecker(game), len(game.Players), _const.MaxPlayers)
		tools.SendMsg(player.Socket, msg)
		gameMutex.Lock()
	}
}

func gameStartChecker(game structs.Game) bool {
	gameMutex.Lock()
	defer gameMutex.Unlock()
	fmt.Println("Player count in lobby: ", len(game.Players))
	fmt.Println("Is game in lobby? ", game.GameData.StartingPhase)
	return len(game.Players) >= 1 && game.GameData.StartingPhase
}

func gameInfoBroadcaster() {
	for _, player := range clientsMap {
		gameMutex.Unlock()
		gameInfoSender(player.Socket)
		gameMutex.Lock()
	}
}

func playerMover(player structs.Player) {
	msg := tools.JoinMsg(true)
	tools.SendMsg(player.Socket, msg)
}

func isGameNotFull(game structs.Game) bool {
	return len(game.Players) < _const.MaxPlayers
}

func gameInfoSender(client net.Conn) {
	password := _const.Pass
	messageType := _const.GamesInfo

	gameMutex.Lock()
	var gameStrings []string
	for _, game := range gameMap {
		playerCount := len(game.Players)
		isLobby := 0
		if game.GameData.StartingPhase {
			isLobby = 1
		}
		gameString := fmt.Sprintf("%s|%d|%d|%d", game.ID, _const.MaxPlayers, playerCount, isLobby)
		gameStrings = append(gameStrings, gameString)
	}

	gameMutex.Unlock()
	message := strings.Join(gameStrings, ";")
	messageLength := fmt.Sprintf("%03d", len(message))
	finalMessage := password + messageLength + messageType + message + "\n"
	fmt.Println("Sending: ", finalMessage)
	gameMutex.Lock()
	tools.SendMsg(client, finalMessage)
	gameMutex.Unlock()
}

func nickCreator(client net.Conn, message string) bool {
	messageType := message[len(_const.Pass)+_const.FormatLen : len(_const.Pass)+_const.FormatLen+_const.CmdLength]
	if messageType == "nick" {
		clientsMap[client] = structs.Player{
			Nick:   message[len(_const.Pass)+_const.FormatLen+_const.CmdLength:],
			Socket: client,
		}
		return true
	} else {
		return false
	}
}

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

func msgValidator(message string) bool {
	if len(message) < (len(_const.Pass) + _const.CmdLength + _const.FormatLen) {
		return false
	}

	password := message[:len(_const.Pass)]

	if password != _const.Pass {
		fmt.Printf("Received password: %s, System password: %s\n", password, _const.Pass)
		return false
	}

	lengthStr := message[len(_const.Pass) : len(_const.Pass)+_const.FormatLen]
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return false
	}
	if length != len(message)-len(_const.Pass)-_const.FormatLen-_const.CmdLength {
		fmt.Printf("Length from message: %d, calculated length: %d\n", length, len(message)-len(_const.Pass)-_const.FormatLen-_const.CmdLength)
		return false
	}
	return true
}
