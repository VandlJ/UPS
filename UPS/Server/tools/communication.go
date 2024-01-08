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

func KillerMsgSender2(client net.Conn) {
	pass := _const.Pass
	cmd := _const.Kill2

	msg := fmt.Sprintf("%s%03d%s\n", pass, 0, cmd)

	SendMsg(client, msg)
}

func KillerMsgSender(client net.Conn) {
	pass := _const.Pass
	cmd := _const.Kill

	msg := fmt.Sprintf("%s%03d%s\n", pass, 0, cmd)

	SendMsg(client, msg)
}

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

func CreateCancelMsg() string {
	pass := _const.Pass
	cmd := _const.Stop

	msg := fmt.Sprintf("%s%03d%s\n", pass, 0, cmd)

	return msg
}

func CreatePingMsg() string {
	pass := _const.Pass
	cmd := _const.Ping

	msg := fmt.Sprintf("%s%03d%s\n", pass, 0, cmd)

	return msg
}

func PlayerActionMsg(game structs.Game, player structs.Player) string {
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

func NextRoundMsg(game structs.Game, player structs.Player) string {
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

func EndMsg(game structs.Game) string {
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

func InitMsg(game structs.Game, player structs.Player) string {
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

func JoinMsg(success bool) string {
	pass := _const.Pass
	cmd := _const.GameJoin

	successStr := "0"

	if success {
		successStr = "1"
	}

	msg := fmt.Sprintf("%s%03d%s%s\n", pass, len(successStr), cmd, successStr)

	return msg
}

func GameReady(canBeStarted bool, currPlayers int, maxPlayers int) string {
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
