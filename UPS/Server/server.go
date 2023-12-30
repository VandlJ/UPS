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

var clientsMapMutex sync.Mutex
var gameMapMutex sync.Mutex

func main() {
	initialGameMap()

	socket, err := net.Listen(constants.ConnType, constants.ConnHost+":"+constants.ConnPort)

	if err != nil {
		fmt.Println("LISTENING - ERR", err.Error())
		os.Exit(1)
	} else {
		fmt.Println("LISTENING - OK")
	}
	defer socket.Close()

	fmt.Println("Server is running...")

	for {
		client, err := socket.Accept()
		if err != nil {
			fmt.Println("ACCEPT - ERR", err.Error())
			return
		}

		fmt.Println("Client " + client.RemoteAddr().String() + " connected.")

		go handleConnection(client)
	}
}

func initialGameMap() {
	gameMapMutex.Lock()
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
	gameMapMutex.Unlock()
}

func handleConnection(client net.Conn) {
	reader := bufio.NewReader(client)

	for {
		readBuffer, err := reader.ReadBytes('\n')

		if err != nil {
			clientsMapMutex.Lock()
			fmt.Println("Killing ", clientsMap[client].Nickname)
			clientsMapMutex.Unlock()
			fmt.Println("Client disconnected: ", client)
			return
		}

		// Converts to string and removes trailing newline chars
		message := strings.TrimRight(string(readBuffer), "\r\n")
		fmt.Println("Message: ", message)

		if isMessageValid(message) {
			fmt.Println("Message structure is valid.")
			handleMessage(message, client)
		} else {
			fmt.Println("Message structure is invalid. Closing connection.")
			return
		}
	}
}

func handleMessage(message string, client net.Conn) {
	clientsMapMutex.Lock()
	if _, exists := clientsMap[client]; !exists && findPlayerBySocket(client) == false {
		if createNickForConnection(client, message) {
			fmt.Println("Client successfully added with nick: ", clientsMap[client].Nickname)
			sendGameInfo(client)
		} else {
			fmt.Println("Identify yourself! Aborting.")
			client.Close()
		}
	} else {
		messageType := message[len(constants.MessageHeader)+constants.MessageLengthFormat : len(constants.MessageHeader)+constants.MessageLengthFormat+constants.MessageTypeLength]
		extractedMessage := message[len(constants.MessageHeader)+constants.MessageLengthFormat+constants.MessageTypeLength:]

		switch messageType {
		case "JOIN":
			joinPlayer(client, message)
		case "INFO":
			printGameMap()
		case "PLAY":
			startGame(client, extractedMessage)
		case "TURN":
			receiveGameChoice(client, extractedMessage)
		default:
			fmt.Println("Unknown command: ", messageType)
		}
	}
	clientsMapMutex.Unlock()
}

func receiveGameChoice(client net.Conn, message string) {
	player := findPlayerBySocketReturn(client)

	if message != "STAND" && message != "HIT" {
		fmt.Println("Invalid game choice!")
		return
	}

	if player == nil {
		fmt.Println("Could find specified player. Aborting.")
		return
	}

	gameID := findGameWithPlayer(*player).ID
	game, ok := gameMap[gameID]
	if ok {
		gameMapMutex.Lock()
		// game.GameData.RoundIndex += 1
		// gameMap[gameID] = game
		playerMadeMove(&game, *player, message, gameID)
		// gameMap[gameID] = game
		if game.GameData.IsLobby {
			print("Game has ended.")
			updateGameInfoInOtherClients()
		}
		gameMapMutex.Unlock()
	}
}

func playerMadeMove(game *structures.Game, player structures.Player, turn string, gameID string) {
	fmt.Printf("%s has played.\n", player.Nickname)
	fmt.Println("Turn was: ", turn)

	fmt.Println("Active players: ", game.GameData.ActivePlayers)

	if turn == "HIT" {
		if game.GameData.Stand[player] == false {
			deckLength := len(game.GameData.Deck.Cards)
			fmt.Printf("HIT - Deck length: %d\n", deckLength)

			newCards := dealCards(&game.GameData.Deck, 1)
			fmt.Println("New cards:", newCards)

			// Create a new slice to store the player's hand to prevent sharing references
			existingHand := structures.Hand{
				Cards: make([]structures.Card, len(game.GameData.PlayerHands[player].Cards)),
			}
			copy(existingHand.Cards, game.GameData.PlayerHands[player].Cards)

			// Append new cards to the existing hand's cards
			existingHand.Cards = append(existingHand.Cards, newCards.Cards...)

			game.GameData.PlayerHands[player] = existingHand

			fmt.Println("HIT - Hand: ", game.GameData.PlayerHands[player])

			calculatePlayerHandValue(&game.GameData, player)

			game.GameData.RoundIndex += 1

			gameMap[gameID] = *game

			for _, player := range gameMap[gameID].Players {
				messageToClients := utils.GameTurnInfo(*game, player)
				player.Socket.Write([]byte(messageToClients))
			}

			fmt.Println("HIT větev")
			fmt.Println("RoundIndex: ", game.GameData.RoundIndex)
			fmt.Println("PlayerCount: ", len(game.Players))
			fmt.Println("Active PlayerCount: ", game.GameData.ActivePlayers)

			if game.GameData.RoundIndex%game.GameData.ActivePlayers == 0 {
				fmt.Println("HIT větev")
				fmt.Println("Every player has played.")
				for _, player := range gameMap[gameID].Players {
					messageToClients := utils.GameNextRound(*game, player)
					fmt.Println("Tuuu: ", messageToClients)
					player.Socket.Write([]byte(messageToClients))
				}
			}
		} else {
			for _, player := range gameMap[gameID].Players {
				messageToClients := utils.GameTurnInfo(*game, player)
				player.Socket.Write([]byte(messageToClients))

				game.GameData.RoundIndex += 1

				gameMap[gameID] = *game
			}
		}
	} else if turn == "STAND" {
		game.GameData.Stand[player] = true
		if game.GameData.Stand[player] == true {
			fmt.Println("Stand status: ", game.GameData.Stand)

			game.GameData.ActivePlayers -= 1
			fmt.Println("Active players in STAND: ", game.GameData.ActivePlayers)
			gameMap[gameID] = *game

			if game.GameData.ActivePlayers == 0 {
				fmt.Println("TAK A KONČÍME")

				// Použití funkce a přidání vítěze do pole Winners
				winner := whoIsTheWinner(game.GameData)
				if winner != nil {
					fmt.Println("Winner is: ", winner)
					game.GameData.Winners = append(game.GameData.Winners, winner.Nickname)
				}

				for _, player := range gameMap[gameID].Players {
					messageToClients := utils.GameEnd(*game)
					fmt.Println("KONEC: ", messageToClients)
					player.Socket.Write([]byte(messageToClients))
				}
			} else {
				fmt.Println("STAND větev")
				fmt.Println("RoundIndex: ", game.GameData.RoundIndex)
				fmt.Println("PlayerCount: ", len(game.Players))

				if game.GameData.RoundIndex%game.GameData.ActivePlayers == 0 {
					fmt.Println("STAND větev")
					fmt.Println("Every player has played.")
					for _, player := range gameMap[gameID].Players {
						messageToClients := utils.GameNextRound(*game, player)
						fmt.Println("Tuuu: ", messageToClients)
						player.Socket.Write([]byte(messageToClients))
					}
				}
			}
		}
	} else {
		fmt.Println("Turn choice broke down")
		return
	}
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

func startGame(client net.Conn, message string) {
	player := findPlayerBySocketReturn(client)
	if player == nil {
		fmt.Println("Play not found in any game. Aborting.")
		return
	}
	game := findGameWithPlayer(*player)
	if canGameBeStarted(*game) {
		switchGameToStart(game.ID)
	} else {
		fmt.Println("Could not switch to game - not enough players.")
	}
}

func switchGameToStart(gameID string) {
	fmt.Println("Updatuju lobby: ", gameID)
	gameMapMutex.Lock()
	defer gameMapMutex.Unlock()

	if existingGame, ok := gameMap[gameID]; ok {
		existingGame.GameData.IsLobby = false

		// Inicializace potřebných proměnných pro hru Blackjack
		existingGame.GameData.Deck = createDeck() // Vytvoření balíčku karet pro hru
		fmt.Println("Deck: ")
		fmt.Println(existingGame.GameData.Deck)

		existingGame.GameData.Deck = shuffleDeck(existingGame.GameData.Deck)
		fmt.Println("Shuffled deck: ")
		fmt.Println(existingGame.GameData.Deck)

		existingGame.GameData.PlayerHands = make(map[structures.Player]structures.Hand) // Mapa pro uchování karet hráčů

		fmt.Println(existingGame.GameData.PlayerHands)

		// Distribuce karet hráčům
		for _, player := range existingGame.Players {
			existingGame.GameData.Stand[player] = false
			initialHand := dealCards(&existingGame.GameData.Deck, 2)
			existingGame.GameData.PlayerHands[player] = initialHand
			calculatePlayerHandValue(&existingGame.GameData, player) // Inicializace celkové hodnoty karet v ruce hráče
			fmt.Println(existingGame.GameData.PlayerHandValue[player])
		}

		printPlayerHands(existingGame.GameData.PlayerHands) // Výpis karet pro hráče

		existingGame.GameData.ActivePlayers = len(existingGame.Players)

		gameMap[gameID] = existingGame

		for _, player := range gameMap[gameID].Players {
			messageToClients := utils.GameStartedWithInitInfo(existingGame, player)
			player.Socket.Write([]byte(messageToClients))

			existingGame.GameData.RoundIndex = 0
		}
		return
	}
}

// initPlayerHandValue inicializuje hodnotu gameData.PlayerHandValue[player] jako součet hodnot karet v ruce hráče
func calculatePlayerHandValue(gameData *structures.GameState, player structures.Player) {
	hand := gameData.PlayerHands[player]
	totalValue := 0

	for _, card := range hand.Cards {
		totalValue += card.Value // Předpokládejme, že hodnota karty je dostupná v card.Value
	}

	gameData.PlayerHandValue[player] = totalValue
}

func dealCards(deck *structures.Deck, cardsCount int) structures.Hand {
	var hand structures.Hand

	fmt.Printf("Deck size: %d, Liznuto karet: %d\n", len(deck.Cards), cardsCount)

	if len(deck.Cards) < cardsCount {
		fmt.Println("Balíček má nedostatek karet pro rozdání.")
		return hand
	}

	hand.Cards = deck.Cards[:cardsCount]
	*deck = structures.Deck{Cards: deck.Cards[cardsCount:]} // Aktualizace původního balíčku karet

	fmt.Printf("Deck size: %d, Liznuto karet: %d\n", len(deck.Cards), cardsCount)

	return hand
}

func shuffleDeck(deck structures.Deck) structures.Deck {
	rand.Seed(time.Now().UnixNano()) // Inicializace generátoru náhodných čísel

	numOfCards := len(deck.Cards)
	shuffledDeck := make([]structures.Card, numOfCards)
	perm := rand.Perm(numOfCards)

	for i, j := range perm {
		shuffledDeck[j] = deck.Cards[i]
	}

	return structures.Deck{Cards: shuffledDeck}
}

func createDeck() structures.Deck {
	var deck structures.Deck

	suits := []string{"Hearts", "Diamonds", "Clubs", "Spades"}
	values := []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 10, 10, 10, 11} // Hodnoty karet

	for _, suit := range suits {
		for i := 0; i < len(values); i++ {
			card := structures.Card{Suit: suit, Value: values[i]}
			deck.Cards = append(deck.Cards, card)
		}
	}

	return deck
}

func printPlayerHands(playerHands map[structures.Player]structures.Hand) {
	for player, hand := range playerHands {
		fmt.Printf("Player %v has cards: %v\n", player, hand.Cards)
	}
}

func findGameWithPlayer(player structures.Player) *structures.Game {
	gameMapMutex.Lock()
	defer gameMapMutex.Unlock()
	for _, game := range gameMap {
		for _, p := range game.Players {
			if p == player {
				return &game
			}
		}
	}
	return nil
}

func findPlayerBySocketReturn(client net.Conn) *structures.Player {
	gameMapMutex.Lock()
	defer gameMapMutex.Unlock()
	for _, gameState := range gameMap {
		for _, player := range gameState.Players {
			if player.Socket == client {
				return &player
			}
		}
	}
	return nil
}

func printGameMap() {
	fmt.Printf("Printing games: \n")
	gameMapMutex.Lock()
	for gameID, game := range gameMap {
		fmt.Printf("Game %s isLobby:%b", gameID, game.GameData.IsLobby)
		fmt.Printf("Number of players: %d\n", len(game.Players))
	}
	gameMapMutex.Unlock()
	fmt.Printf("Printing main game: \n")
	for client := range clientsMap {
		fmt.Printf("Client %s with username %s\n", client.RemoteAddr(), clientsMap[client].Nickname)
	}
}

func joinPlayer(client net.Conn, message string) {
	gameName := message[len(constants.MessageHeader)+constants.MessageLengthFormat+constants.MessageTypeLength:]
	gameMapMutex.Lock()
	if game, ok := gameMap[gameName]; ok {
		if isGameEmpty(game) {
			if _, exists := clientsMap[client]; exists {
				playerID := len(game.Players) + 1
				game.Players[playerID] = clientsMap[client]
				fmt.Printf("User %s has joined the game %s\n", clientsMap[client].Nickname, gameName)
				delete(clientsMap, client)
				playerMovedToGameLobby(game.Players[playerID])
				updateGameInfoInOtherClients()
				sendInfoAboutStart(game)
			} else {
				fmt.Println("User not found in clients map.")
			}
		} else {
			fmt.Println("Lobby is not empty.")
		}
	} else {
		fmt.Printf("Lobby %s not found in game map.\n", gameName)
	}
	gameMapMutex.Unlock()
}

func sendInfoAboutStart(game structures.Game) {
	for _, player := range game.Players {
		gameMapMutex.Unlock()
		player.Socket.Write([]byte(utils.CanBeStarted(canGameBeStarted(game), len(game.Players), constants.MaxPlayers)))
		gameMapMutex.Lock()
	}
}

func canGameBeStarted(game structures.Game) bool {
	gameMapMutex.Lock()
	defer gameMapMutex.Unlock()
	return len(game.Players) >= 1 && game.GameData.IsLobby
}

func updateGameInfoInOtherClients() {
	for _, player := range clientsMap {
		gameMapMutex.Unlock()
		sendGameInfo(player.Socket)
		gameMapMutex.Lock()
	}
}

func playerMovedToGameLobby(player structures.Player) {
	player.Socket.Write([]byte(utils.GameJoined(true)))
}

func isGameEmpty(game structures.Game) bool {
	return len(game.Players) < constants.MaxPlayers
}

func sendGameInfo(client net.Conn) {
	password := constants.MessageHeader
	messageType := constants.GamesInfo

	gameMapMutex.Lock()
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

	gameMapMutex.Unlock()
	message := strings.Join(gameStrings, ";")
	messageLength := fmt.Sprintf("%03d", len(message))
	finalMessage := password + messageLength + messageType + message + "\n"
	fmt.Println("Sending: ", finalMessage)
	gameMapMutex.Lock()
	_, err := client.Write([]byte(finalMessage))
	gameMapMutex.Unlock()
	if err != nil {
		return
	}
}

func createNickForConnection(client net.Conn, message string) bool {
	messageType := message[len(constants.MessageHeader)+constants.MessageLengthFormat : len(constants.MessageHeader)+constants.MessageLengthFormat+constants.MessageTypeLength]
	if messageType == "nick" {
		clientsMap[client] = structures.Player{
			Nickname: message[len(constants.MessageHeader)+constants.MessageLengthFormat+constants.MessageTypeLength:],
			Socket:   client,
		}
		return true
	} else {
		return false
	}
}

func findPlayerBySocket(client net.Conn) bool {
	gameMapMutex.Lock()
	defer gameMapMutex.Unlock()
	for _, gameState := range gameMap {
		for _, player := range gameState.Players {
			if player.Socket == client {
				return true
			}
		}
	}
	return false
}

func isMessageValid(message string) bool {
	if len(message) < (len(constants.MessageHeader) + constants.MessageTypeLength + constants.MessageLengthFormat) {
		return false
	}

	// Password
	password := message[:len(constants.MessageHeader)]

	if password != constants.MessageHeader {
		fmt.Printf("Received password: %s, System password: %s\n", password, constants.MessageHeader)
		return false
	}

	// Message length
	lengthStr := message[len(constants.MessageHeader) : len(constants.MessageHeader)+constants.MessageLengthFormat]
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return false
	}

	// Is message length valid?
	if length != len(message)-len(constants.MessageHeader)-constants.MessageLengthFormat-constants.MessageTypeLength {
		fmt.Printf("Length from message: %d, calculated length: %s\n", length, len(message)-len(constants.MessageHeader)-constants.MessageLengthFormat-constants.MessageTypeLength)
		return false
	}

	return true
}
