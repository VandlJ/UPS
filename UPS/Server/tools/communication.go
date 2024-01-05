package tools

import (
	"Server/const"
	"Server/structs"
	"fmt"
	"net"
	"strings"
)

func SendMsg(socket net.Conn, msg string) {
	socket.Write([]byte(msg))
}

func PingPongIntervalMsg() string {
	password := _const.Pass
	msgType := _const.PingIntervalSet
	msg := fmt.Sprintf("%s%03d%s|%d\n", password, 0, msgType, _const.PingInterval)
	return msg
}

func CreateCancelMessage() string {
	magic := _const.Pass
	messageType := _const.Cancel
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
	players := getPlayerNicknameWithPoints(game, player)
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
	players := getPlayerNicknameWithPoints(game, player)
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
	players := getPlayerNicknameWithPoints(game, player)
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

func getPlayerNicknameWithPoints(game structs.Game, player structs.Player) string {
	return fmt.Sprintf("%s:%d", player.Nick, game.GameData.PlayerHandValue[player])
}
