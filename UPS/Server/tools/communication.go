package tools

import (
	"Server/const"
	"Server/structs"
	"fmt"
	"net"
	"strings"
)

func SendMsg(socket net.Conn, msg string) {
	_, err := socket.Write([]byte(msg))
	if err != nil {
		return
	}
}

func PlayerStateSender(game structs.Game, checkedPlayer structs.Player, state string) {
	pass := _const.Pass
	msgType := _const.State
	playerState := state
	msgLength := fmt.Sprintf("%03d", len(checkedPlayer.Nick))
	msg := pass + msgLength + msgType + checkedPlayer.Nick + "|" + playerState + "\n"

	for _, player := range game.Players {
		SendMsg(player.Socket, msg)
	}
}

func CreateResendStateMessage(game *structs.Game, player structs.Player) string {
	password := _const.Pass
	messageType := _const.RetrieveState
	status := 0 // did not play
	if game.GameData.HasPlayed[player] == true {
		status = 1 // already played
	}
	// players := getPlayerNicknameWithPoints(*game, player)
	players := getPlayerNicks(*game)
	fmt.Println("Players: ", players)
	playerHand := game.GameData.PlayerHands[player]
	playerCardsString := getPlayerCardsString(playerHand.Cards)
	playerHandValue := game.GameData.PlayerHandValue[player]
	messageBody := fmt.Sprintf("%s|%s|%d|%d", players, playerCardsString, playerHandValue, status)
	fmt.Println("msg body: ", messageBody)
	finalMessage := fmt.Sprintf("%s%03d%s%s\n", password, len(messageBody), messageType, messageBody)
	fmt.Println("STATE: ", finalMessage)
	return finalMessage
}

func PingPongIntervalMsg() string {
	password := _const.Pass
	msgType := _const.PingIntervalSet
	msg := fmt.Sprintf("%s%03d%s|%d\n", password, 0, msgType, _const.PingInterval)
	return msg
}

func CreateCancelMessage() string {
	magic := _const.Pass
	messageType := _const.Stop
	message := fmt.Sprintf("%s%03d%s\n", magic, 0, messageType)
	return message
}

func CreatePingMessage() string {
	magic := _const.Pass
	messageType := _const.Ping
	message := fmt.Sprintf("%s%03d%s\n", magic, 0, messageType)
	return message
}

func PlayerActionMsg(game structs.Game, player structs.Player) string {
	password := _const.Pass
	messageType := _const.GameTurn
	// players := getPlayerNicknameWithPoints(game, player)
	players := getPlayerNicks(game)
	playerHand := game.GameData.PlayerHands[player]
	playerCardsString := getPlayerCardsString(playerHand.Cards)
	playerHandValue := game.GameData.PlayerHandValue[player]
	messageBody := fmt.Sprintf("%s|%s|%d", players, playerCardsString, playerHandValue)
	finalMessage := fmt.Sprintf("%s%03d%s%s\n", password, len(messageBody), messageType, messageBody)
	return finalMessage
}

func NextRoundMsg(game structs.Game, player structs.Player) string {
	password := _const.Pass
	messageType := _const.GameNextRound
	// players := getPlayerNicknameWithPoints(game, player)
	players := getPlayerNicks(game)
	playerHand := game.GameData.PlayerHands[player]
	playerCardsString := getPlayerCardsString(playerHand.Cards)
	playerHandValue := game.GameData.PlayerHandValue[player]
	messageBody := fmt.Sprintf("%s|%s|%d", players, playerCardsString, playerHandValue)
	finalMessage := fmt.Sprintf("%s%03d%s%s\n", password, len(messageBody), messageType, messageBody)
	return finalMessage
}

func EndMsg(game structs.Game) string {
	password := _const.Pass
	messageType := _const.GameEnd

	var winnersString strings.Builder
	for _, winner := range game.GameData.Winners {
		winnersString.WriteString(winner)
		winnersString.WriteString(";")
	}
	winnersStringStripped := strings.TrimSuffix(winnersString.String(), ";")

	messageBody := winnersStringStripped
	finalMessage := fmt.Sprintf("%s%03d%s%s\n", password, len(messageBody), messageType, messageBody)
	return finalMessage
}

func InitMsg(game structs.Game, player structs.Player) string {
	password := _const.Pass
	messageType := _const.GameStart
	//players := getPlayerNicknameWithPoints(game, player)
	players := getPlayerNicks(game)
	playerHand := game.GameData.PlayerHands[player]
	playerCardsString := getPlayerCardsString(playerHand.Cards)
	playerHandValue := game.GameData.PlayerHandValue[player]
	messageBody := fmt.Sprintf("%s|%s|%d", players, playerCardsString, playerHandValue)
	finalMessage := fmt.Sprintf("%s%03d%s%s\n", password, len(messageBody), messageType, messageBody)
	return finalMessage
}

func JoinMsg(success bool) string {
	password := _const.Pass
	messageType := _const.GameJoin
	successStr := "0"
	if success {
		successStr = "1"
	}

	message := fmt.Sprintf("%s%03d%s%s\n", password, len(successStr), messageType, successStr)
	return message
}

func GameReady(canBeStarted bool, currPlayers int, maxPlayers int) string {
	password := _const.Pass
	messageType := _const.GameStartCheck
	canBeStartedStr := "0"
	if canBeStarted {
		canBeStartedStr = "1"
	}

	messageBody := fmt.Sprintf("%s|%d|%d", canBeStartedStr, currPlayers, maxPlayers)
	message := fmt.Sprintf("%s%03d%s%s\n", password, len(messageBody), messageType, messageBody)
	return message
}

func getPlayerCardsString(cards []structs.Card) string {
	cardsString := ""
	for _, card := range cards {
		if cardsString != "" {
			cardsString += ", "
		}
		cardsString += fmt.Sprintf("%s %d", card.Suit, card.Value)
	}
	return cardsString
}

// getPlayerNicks returns a slice of nicknames of all players in the game.
func getPlayerNicks(game structs.Game) string {
	var nicknames []string
	var players string

	for _, player := range game.Players {
		nicknames = append(nicknames, player.Nick)
	}

	for i, nick := range nicknames {
		if i > 0 {
			players += ";"
		}
		players += fmt.Sprintf("%s", nick)
	}

	return players
}
