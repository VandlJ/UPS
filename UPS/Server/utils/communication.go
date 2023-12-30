package utils

import (
	"Server/constants"
	"Server/structures"
	"fmt"
	"strings"
)

func PlayerActionMsg(game structures.Game, player structures.Player) string {
	password := constants.Password
	messageType := constants.GameTurn
	players := getPlayerNicknameWithPoints(game, player)
	playerHand := game.GameData.PlayerHands[player]
	playerCardsString := getPlayerCardsString(playerHand.Cards)
	playerHandValue := game.GameData.PlayerHandValue[player]
	messageBody := fmt.Sprintf("%s|%s|%d", players, playerCardsString, playerHandValue)
	finalMessage := fmt.Sprintf("%s%03d%s%s\n", password, len(messageBody), messageType, messageBody)
	return finalMessage
}

func NextRoundMsg(game structures.Game, player structures.Player) string {
	password := constants.Password
	messageType := constants.GameNextRound
	players := getPlayerNicknameWithPoints(game, player)
	playerHand := game.GameData.PlayerHands[player]
	playerCardsString := getPlayerCardsString(playerHand.Cards)
	playerHandValue := game.GameData.PlayerHandValue[player]
	messageBody := fmt.Sprintf("%s|%s|%d", players, playerCardsString, playerHandValue)
	finalMessage := fmt.Sprintf("%s%03d%s%s\n", password, len(messageBody), messageType, messageBody)
	return finalMessage
}

func EndMsg(game structures.Game) string {
	password := constants.Password
	messageType := constants.GameEnd

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

func InitMsg(game structures.Game, player structures.Player) string {
	password := constants.Password
	messageType := constants.GameStart
	players := getPlayerNicknameWithPoints(game, player)
	playerHand := game.GameData.PlayerHands[player]
	playerCardsString := getPlayerCardsString(playerHand.Cards)
	playerHandValue := game.GameData.PlayerHandValue[player]
	messageBody := fmt.Sprintf("%s|%s|%d", players, playerCardsString, playerHandValue)
	finalMessage := fmt.Sprintf("%s%03d%s%s\n", password, len(messageBody), messageType, messageBody)
	return finalMessage
}

func JoinMsg(success bool) string {
	password := constants.Password
	messageType := constants.GameJoin
	successStr := "0"
	if success {
		successStr = "1"
	}

	message := fmt.Sprintf("%s%03d%s%s\n", password, len(successStr), messageType, successStr)
	return message
}

func GameReady(canBeStarted bool, currPlayers int, maxPlayers int) string {
	password := constants.Password
	messageType := constants.GameStartCheck
	canBeStartedStr := "0"
	if canBeStarted {
		canBeStartedStr = "1"
	}

	messageBody := fmt.Sprintf("%s|%d|%d", canBeStartedStr, currPlayers, maxPlayers)
	message := fmt.Sprintf("%s%03d%s%s\n", password, len(messageBody), messageType, messageBody)
	return message
}

func getPlayerCardsString(cards []structures.Card) string {
	cardsString := ""
	for _, card := range cards {
		if cardsString != "" {
			cardsString += ", "
		}
		cardsString += fmt.Sprintf("%s %d", card.Suit, card.Value)
	}
	return cardsString
}

func getPlayerNicknameWithPoints(game structures.Game, player structures.Player) string {
	return fmt.Sprintf("%s:%d", player.Nick, game.GameData.PlayerHandValue[player])
}
