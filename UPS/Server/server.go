package main

import (
	"Server/constants"
	"Server/structures"
	"Server/utils"
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

var clientsMap = make(map[net.Conn]structures.Player)
var gameMap = make(map[string]structures.Game)

var clientsMutex sync.Mutex
var gameMutex sync.Mutex

func main() {
	gameMapInit()

	socket, err := net.Listen(constants.ConnType, constants.ConnHost+":"+constants.ConnPort)

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

		go connHandler(client)
	}
}

func gameMapInit() {
	gameMutex.Lock()
	for i := 1; i <= constants.GameRoomsCount; i++ {
		gameID := fmt.Sprintf("game%d", i)
		gameMap[gameID] = structures.Game{
			ID:      gameID,
			Players: make(map[int]structures.Player),
			GameData: structures.GameState{
				IsLobby:         true,
				PlayerHandValue: make(map[structures.Player]int),
				Stand:           make(map[structures.Player]bool),
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
			clientsMutex.Unlock()
			fmt.Println("Client disconnected: ", client)
			return
		}

		// Converts to string and removes trailing newline chars
		msg := strings.TrimRight(string(readBuff), "\r\n")
		fmt.Println("Message: ", msg)

		if msgValidator(msg) {
			fmt.Println("Message structure is valid.")
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
		msgType := msg[len(constants.Password)+constants.MessageLengthFormat : len(constants.Password)+constants.MessageLengthFormat+constants.MessageTypeLength]
		msgContent := msg[len(constants.Password)+constants.MessageLengthFormat+constants.MessageTypeLength:]

		switch msgType {
		case "JOIN":
			playerJoiner(client, msg)
		case "PLAY":
			startGame(client)
		case "TURN":
			playerActionReceiver(client, msgContent)
		default:
			fmt.Println("Unknown command: ", msgType)
		}
	}
	clientsMutex.Unlock()
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
		if game.GameData.IsLobby {
			fmt.Println("Game has ended")
			gameInfoBroadcaster()
		}
		gameMutex.Unlock()
	}
}

func playerActionHandler(game *structures.Game, player structures.Player, turn string, gameID string) {
	fmt.Printf("%s has played %s.\n", player.Nick, turn)
	fmt.Println("Active players: ", game.GameData.ActivePlayers)

	if turn == "HIT" {
		if game.GameData.Stand[player] == false {
			deckSize := len(game.GameData.Deck.Cards)
			fmt.Printf("HIT - Deck size: %d\n", deckSize)

			newCard := cardDealer(&game.GameData.Deck, 1)
			fmt.Println("New cards:", newCard)

			existingHand := structures.Hand{
				Cards: make([]structures.Card, len(game.GameData.PlayerHands[player].Cards)),
			}
			copy(existingHand.Cards, game.GameData.PlayerHands[player].Cards)

			existingHand.Cards = append(existingHand.Cards, newCard.Cards...)

			game.GameData.PlayerHands[player] = existingHand

			fmt.Println("HIT - Hand: ", game.GameData.PlayerHands[player])

			handValueCalculator(&game.GameData, player)

			game.GameData.RoundIndex += 1

			gameMap[gameID] = *game

			for _, player := range gameMap[gameID].Players {
				msgToBroadcast := utils.PlayerActionMsg(*game, player)
				fmt.Println("Message to broadcast: ", msgToBroadcast)
				_, err := player.Socket.Write([]byte(msgToBroadcast))
				if err != nil {
					return
				}
			}

			fmt.Println("HIT branch")
			fmt.Println("RoundIndex: ", game.GameData.RoundIndex)
			fmt.Println("Active PlayerCount: ", game.GameData.ActivePlayers)

			if game.GameData.RoundIndex%game.GameData.ActivePlayers == 0 {
				fmt.Println("HIT branch")
				fmt.Println("Every player has played")
				for _, player := range gameMap[gameID].Players {
					msgToBroadcast := utils.NextRoundMsg(*game, player)
					fmt.Println("Message to broadcast: ", msgToBroadcast)
					_, err := player.Socket.Write([]byte(msgToBroadcast))
					if err != nil {
						return
					}
				}
			}
		} else {
			for _, player := range gameMap[gameID].Players {
				msgToBroadcast := utils.PlayerActionMsg(*game, player)
				fmt.Println("Message to broadcast: ", msgToBroadcast)
				_, err := player.Socket.Write([]byte(msgToBroadcast))
				if err != nil {
					return
				}

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
					msgToBroadcast := utils.EndMsg(*game)
					fmt.Println("Message to broadcast: ", msgToBroadcast)
					_, err := player.Socket.Write([]byte(msgToBroadcast))
					if err != nil {
						return
					}
				}

				playerDisconnector(game)
				game.GameData.IsLobby = true
				gameMap[gameID] = *game

			} else {
				fmt.Println("STAND branch")
				fmt.Println("RoundIndex: ", game.GameData.RoundIndex)
				fmt.Println("Active PlayerCount: ", game.GameData.ActivePlayers)

				if game.GameData.RoundIndex%game.GameData.ActivePlayers == 0 {
					fmt.Println("STAND branch")
					fmt.Println("Every player has played")
					for _, player := range gameMap[gameID].Players {
						msgToBroadcast := utils.NextRoundMsg(*game, player)
						fmt.Println("Message to broadcast: ", msgToBroadcast)
						_, err := player.Socket.Write([]byte(msgToBroadcast))
						if err != nil {
							return
						}
					}
				}
			}
		}
	} else {
		fmt.Println("Player action handler broke down")
		return
	}
}

func playerDisconnector(game *structures.Game) {
	for _, player := range game.Players {
		clientsMap[player.Socket] = player
	}
	game.Players = make(map[int]structures.Player)
}

func whoIsTheWinner(gameData structures.GameState) *structures.Player {
	var winner *structures.Player
	highestScore := 0

	for player, score := range gameData.PlayerHandValue {
		if score > highestScore && score <= 21 {
			highestScore = score
			playerCopy := player
			winner = &playerCopy
		}
	}
	fmt.Println("Winner: ", winner)
	return winner
}

func startGame(client net.Conn) {
	player := clientConnReturn(client)
	if player == nil {
		fmt.Println("Player not found.")
		return
	}
	game := playerGameFinder(*player)
	if gameStartChecker(*game) {
		gameStartHandler(game.ID)
	} else {
		fmt.Println("Could not switch to game - not enough players.")
	}
}

func gameStartHandler(gameID string) {
	fmt.Println("Starting game ", gameID)
	gameMutex.Lock()
	defer gameMutex.Unlock()

	if existingGame, ok := gameMap[gameID]; ok {
		existingGame.GameData.IsLobby = false

		existingGame.GameData.Deck = createDeck()
		fmt.Println("Deck created")

		existingGame.GameData.Deck = shuffleDeck(existingGame.GameData.Deck)
		fmt.Println("Deck shuffled")

		existingGame.GameData.PlayerHands = make(map[structures.Player]structures.Hand)
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
			msgToBroadcast := utils.InitMsg(existingGame, player)
			_, err := player.Socket.Write([]byte(msgToBroadcast))
			if err != nil {
				return
			}
			existingGame.GameData.RoundIndex = 0
		}
		return
	}
}

func handValueCalculator(gameData *structures.GameState, player structures.Player) {
	hand := gameData.PlayerHands[player]
	totalValue := 0

	for _, card := range hand.Cards {
		totalValue += card.Value
	}
	gameData.PlayerHandValue[player] = totalValue
}

func cardDealer(deck *structures.Deck, cardsCount int) structures.Hand {
	var hand structures.Hand

	fmt.Printf("Deck size: %d\nCards dealt: %d\n", len(deck.Cards), cardsCount)

	if len(deck.Cards) < cardsCount {
		fmt.Println("Not enough cards in deck")
		return hand
	}

	hand.Cards = deck.Cards[:cardsCount]
	*deck = structures.Deck{Cards: deck.Cards[cardsCount:]}

	fmt.Printf("Deck size: %d\n", len(deck.Cards))

	return hand
}

func shuffleDeck(deck structures.Deck) structures.Deck {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))

	numOfCards := len(deck.Cards)
	shuffledDeck := make([]structures.Card, numOfCards)
	perm := random.Perm(numOfCards)

	for i, j := range perm {
		shuffledDeck[j] = deck.Cards[i]
	}

	return structures.Deck{Cards: shuffledDeck}
}

func createDeck() structures.Deck {
	var deck structures.Deck

	suits := []string{"Hearts", "Diamonds", "Clubs", "Spades"}
	values := []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 10, 10, 10, 11}

	for _, suit := range suits {
		for i := 0; i < len(values); i++ {
			card := structures.Card{Suit: suit, Value: values[i]}
			deck.Cards = append(deck.Cards, card)
		}
	}
	return deck
}

func playerHandsPrinter(playerHands map[structures.Player]structures.Hand) {
	for player, hand := range playerHands {
		fmt.Printf("Player %v has cards: %v\n", player, hand.Cards)
	}
}

func playerGameFinder(player structures.Player) *structures.Game {
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

func clientConnReturn(client net.Conn) *structures.Player {
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
	gameName := msg[len(constants.Password)+constants.MessageLengthFormat+constants.MessageTypeLength:]
	gameMutex.Lock()
	if game, ok := gameMap[gameName]; ok {
		if isGameNotFull(game) {
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
			}
		} else {
			fmt.Println("Lobby is not empty.")
		}
	} else {
		fmt.Printf("Lobby %s not found in game map.\n", gameName)
	}
	gameMutex.Unlock()
}

func startInfoSender(game structures.Game) {
	for _, player := range game.Players {
		gameMutex.Unlock()
		_, err := player.Socket.Write([]byte(utils.GameReady(gameStartChecker(game), len(game.Players), constants.MaxPlayers)))
		if err != nil {
			return
		}
		gameMutex.Lock()
	}
}

func gameStartChecker(game structures.Game) bool {
	gameMutex.Lock()
	defer gameMutex.Unlock()
	fmt.Println("Player count in lobby: ", len(game.Players))
	fmt.Println("Is game in lobby? ", game.GameData.IsLobby)
	return len(game.Players) >= 1 && game.GameData.IsLobby
}

func gameInfoBroadcaster() {
	for _, player := range clientsMap {
		gameMutex.Unlock()
		gameInfoSender(player.Socket)
		gameMutex.Lock()
	}
}

func playerMover(player structures.Player) {
	_, err := player.Socket.Write([]byte(utils.JoinMsg(true)))
	if err != nil {
		return
	}
}

func isGameNotFull(game structures.Game) bool {
	return len(game.Players) < constants.MaxPlayers
}

func gameInfoSender(client net.Conn) {
	password := constants.Password
	messageType := constants.GamesInfo

	gameMutex.Lock()
	var gameStrings []string
	for _, game := range gameMap {
		playerCount := len(game.Players)
		isLobby := 0
		if game.GameData.IsLobby {
			isLobby = 1
		}
		gameString := fmt.Sprintf("%s|%d|%d|%d", game.ID, constants.MaxPlayers, playerCount, isLobby)
		gameStrings = append(gameStrings, gameString)
	}

	gameMutex.Unlock()
	message := strings.Join(gameStrings, ";")
	messageLength := fmt.Sprintf("%03d", len(message))
	finalMessage := password + messageLength + messageType + message + "\n"
	fmt.Println("Sending: ", finalMessage)
	gameMutex.Lock()
	_, err := client.Write([]byte(finalMessage))
	gameMutex.Unlock()
	if err != nil {
		return
	}
}

func nickCreator(client net.Conn, message string) bool {
	messageType := message[len(constants.Password)+constants.MessageLengthFormat : len(constants.Password)+constants.MessageLengthFormat+constants.MessageTypeLength]
	if messageType == "nick" {
		clientsMap[client] = structures.Player{
			Nick:   message[len(constants.Password)+constants.MessageLengthFormat+constants.MessageTypeLength:],
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
	if len(message) < (len(constants.Password) + constants.MessageTypeLength + constants.MessageLengthFormat) {
		return false
	}

	password := message[:len(constants.Password)]

	if password != constants.Password {
		fmt.Printf("Received password: %s, System password: %s\n", password, constants.Password)
		return false
	}

	lengthStr := message[len(constants.Password) : len(constants.Password)+constants.MessageLengthFormat]
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return false
	}
	if length != len(message)-len(constants.Password)-constants.MessageLengthFormat-constants.MessageTypeLength {
		fmt.Printf("Length from message: %d, calculated length: %d\n", length, len(message)-len(constants.Password)-constants.MessageLengthFormat-constants.MessageTypeLength)
		return false
	}
	return true
}
