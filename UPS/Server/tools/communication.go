package tools

import (
	"Server/const"
	"Server/structs"
	"fmt"
	"net"
	"strings"
)

// SendMsg sends a message via the provided socket connection
func SendMsg(socket net.Conn, msg string) {
	_, err := socket.Write([]byte(msg))
	if err != nil {
		return
	}
}

// KillerMsgSender2 sends a KIL2 message to the client
func KillerMsgSender2(client net.Conn) {
	pass := _const.Pass
	cmd := _const.Kill2

	msg := fmt.Sprintf("%s%03d%s\n", pass, 0, cmd)

	SendMsg(client, msg)
}

// KillerMsgSender sends a KILL message to the client
func KillerMsgSender(client net.Conn) {
	pass := _const.Pass
	cmd := _const.Kill

	msg := fmt.Sprintf("%s%03d%s\n", pass, 0, cmd)

	SendMsg(client, msg)
}

// PlayerStateSender sends player state to all players in the game
func PlayerStateSender(game structs.Game, checkedPlayer structs.Player, state string) {
	pass := _const.Pass
	cmd := _const.State

	playerState := state
	msgLen := fmt.Sprintf("%03d", len(checkedPlayer.Nick))

	msg := pass + msgLen + cmd + checkedPlayer.Nick + "|" + playerState + "\n"

	for _, player := range game.Players {
		SendMsg(player.Socket, msg)
	}
}

// CreateReconnectMsg creates a message for player reconnection
func CreateReconnectMsg(game *structs.Game, player structs.Player) string {
	pass := _const.Pass
	cmd := _const.RetrieveState

	status := 0 // did not play
	if game.GameData.HasPlayed[player] == true {
		status = 1 // already played
	}

	players := getPlayerNicks(*game)
	playerHand := game.GameData.PlayerHands[player]
	playerCardsString := getPlayerCardsString(playerHand.Cards)
	playerHandValue := game.GameData.PlayerHandValue[player]

	msgCnt := fmt.Sprintf("%s|%s|%d|%d", players, playerCardsString, playerHandValue, status)

	msg := fmt.Sprintf("%s%03d%s%s\n", pass, len(msgCnt), cmd, msgCnt)

	return msg
}

// CreateStopMsg creates a STOP message
func CreateStopMsg() string {
	pass := _const.Pass
	cmd := _const.Stop

	msg := fmt.Sprintf("%s%03d%s\n", pass, 0, cmd)

	return msg
}

// CreatePingMsg creates a PING message
func CreatePingMsg() string {
	pass := _const.Pass
	cmd := _const.Ping

	msg := fmt.Sprintf("%s%03d%s\n", pass, 0, cmd)

	return msg
}

// CreateTurnMsg creates a message for a player's turn in the game
func CreateTurnMsg(game structs.Game, player structs.Player) string {
	pass := _const.Pass
	cmd := _const.GameTurn

	players := getPlayerNicks(game)
	playerHand := game.GameData.PlayerHands[player]
	playerCardsString := getPlayerCardsString(playerHand.Cards)
	playerHandValue := game.GameData.PlayerHandValue[player]

	msgCnt := fmt.Sprintf("%s|%s|%d", players, playerCardsString, playerHandValue)

	msg := fmt.Sprintf("%s%03d%s%s\n", pass, len(msgCnt), cmd, msgCnt)

	return msg
}

// CreateNextMsg creates a message for the next round in the game
func CreateNextMsg(game structs.Game, player structs.Player) string {
	pass := _const.Pass
	cmd := _const.GameNextRound

	players := getPlayerNicks(game)
	playerStandStatus := game.GameData.Stand[player]
	standStatus := 0
	if playerStandStatus == true {
		standStatus = 1
	}

	playerHand := game.GameData.PlayerHands[player]
	playerCardsString := getPlayerCardsString(playerHand.Cards)
	playerHandValue := game.GameData.PlayerHandValue[player]

	msgCnt := fmt.Sprintf("%s|%s|%d|%d", players, playerCardsString, playerHandValue, standStatus)

	msg := fmt.Sprintf("%s%03d%s%s\n", pass, len(msgCnt), cmd, msgCnt)

	return msg
}

// CreateEndMsg creates a message for the end of the game
func CreateEndMsg(game structs.Game) string {
	pass := _const.Pass
	cmd := _const.GameEnd

	var winnersString strings.Builder
	for _, winner := range game.GameData.Winners {
		winnersString.WriteString(winner)
		winnersString.WriteString(";")
	}
	winnersStringStripped := strings.TrimSuffix(winnersString.String(), ";")

	msgCnt := winnersStringStripped

	msg := fmt.Sprintf("%s%03d%s%s\n", pass, len(msgCnt), cmd, msgCnt)

	return msg
}

// CreateInitMsg creates an initialization message for the start of the game
func CreateInitMsg(game structs.Game, player structs.Player) string {
	pass := _const.Pass
	cmd := _const.GameStart

	players := getPlayerNicks(game)
	playerHand := game.GameData.PlayerHands[player]
	playerCardsString := getPlayerCardsString(playerHand.Cards)
	playerHandValue := game.GameData.PlayerHandValue[player]

	msgCnt := fmt.Sprintf("%s|%s|%d", players, playerCardsString, playerHandValue)

	msg := fmt.Sprintf("%s%03d%s%s\n", pass, len(msgCnt), cmd, msgCnt)

	return msg
}

// CreateJoinMsg creates a JOIN message for game participation
func CreateJoinMsg(success bool) string {
	pass := _const.Pass
	cmd := _const.GameJoin

	successStr := "0"

	if success {
		successStr = "1"
	}

	msg := fmt.Sprintf("%s%03d%s%s\n", pass, len(successStr), cmd, successStr)

	return msg
}

// CreateCheckMsg creates a message to check if the game can be started
func CreateCheckMsg(canBeStarted bool, currPlayers int, maxPlayers int) string {
	pass := _const.Pass
	cmd := _const.GameStartCheck

	canBeStartedStr := "0"

	if canBeStarted {
		canBeStartedStr = "1"
	}

	msgCnt := fmt.Sprintf("%s|%d|%d", canBeStartedStr, currPlayers, maxPlayers)

	msg := fmt.Sprintf("%s%03d%s%s\n", pass, len(msgCnt), cmd, msgCnt)

	return msg
}

// getPlayerCardsString generates a string representation of player cards
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

// getPlayerNicks generates a string containing player nicknames
func getPlayerNicks(game structs.Game) string {
	var nicks []string
	var players string

	for _, player := range game.Players {
		nicks = append(nicks, player.Nick)
	}

	for i, nick := range nicks {
		if i > 0 {
			players += ";"
		}
		players += fmt.Sprintf("%s", nick)
	}

	return players
}
