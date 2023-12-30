package utils

import (
	"Server/constants"
	"Server/structures"
	"fmt"
)

func GameJoined(success bool) string {
	password := constants.MessageHeader
	messageType := constants.GameJoin
	successStr := "0"
	if success {
		successStr = "1"
	}

	message := fmt.Sprintf("%s%03d%s%s\n", password, len(successStr), messageType, successStr)
	return message
}

func CanBeStarted(canBeStarted bool, currPlayers int, maxPlayers int) string {
	password := constants.MessageHeader
	messageType := constants.GameStartCheck
	canBeStartedStr := "0"
	if canBeStarted {
		canBeStartedStr = "1"
	}

	messageBody := fmt.Sprintf("%s|%d|%d", canBeStartedStr, currPlayers, maxPlayers)
	message := fmt.Sprintf("%s%03d%s%s\n", password, len(messageBody), messageType, messageBody)
	return message
}

func GameStartedWithInitInfo(game structures.Game, player structures.Player) string {
	password := constants.MessageHeader
	messageType := constants.GameStart
	players := getPlayerNicknameWithPoints(game, player)
	playerHand := game.GameData.PlayerHands[player]
	playerCardsString := getPlayerCardsString(playerHand.Cards)
	playerHandValue := game.GameData.PlayerHandValue[player]
	messageBody := fmt.Sprintf("%s|%s|%d", players, playerCardsString, playerHandValue)
	finalMessage := fmt.Sprintf("%s%03d%s%s\n", password, len(messageBody), messageType, messageBody)
	return finalMessage
}

func GameTurnInfo(game structures.Game, player structures.Player) string {
	password := constants.MessageHeader
	messageType := constants.GameTurn
	players := getPlayerNicknameWithPoints(game, player)
	playerHand := game.GameData.PlayerHands[player]
	playerCardsString := getPlayerCardsString(playerHand.Cards)
	playerHandValue := game.GameData.PlayerHandValue[player]
	messageBody := fmt.Sprintf("%s|%s|%d", players, playerCardsString, playerHandValue)
	finalMessage := fmt.Sprintf("%s%03d%s%s\n", password, len(messageBody), messageType, messageBody)
	return finalMessage
}

func GameNextRound(game structures.Game, player structures.Player) string {
	password := constants.MessageHeader
	messageType := constants.GameNextRound
	players := getPlayerNicknameWithPoints(game, player)
	playerHand := game.GameData.PlayerHand[player]
	playerCardsString := getPlayerCardsString(playerHand.Cards)
	playerHandValue := game.GameData.PlayerHandValue[player]
	messageBody := fmt.Sprintf("%s|%s|%d", players, playerCardsString, playerHandValue)
	finalMessage := fmt.Sprintf("%s%03d%s%s\n", password, len(messageBody), messageType, messageBody)
	return finalMessage
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
	return fmt.Sprintf("%s:%d", player.Nickname, game.GameData.PlayerHandValue[player])
}
