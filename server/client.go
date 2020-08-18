package main

import (
	"github.com/mikeflynn/the-fearsome-five/shared"
)

func clientMsgRouter(idx *Index, conn *shared.Conn, message *shared.Message) bool {
	switch message.Action {
	case "systemReport":
		body, err := message.GetBodyJSON()

		if err == nil {
			if client, yes := idx.clients[conn]; yes {
				if val, ok := body["os"]; ok {
					client.OS = val
				}

				if val, ok := body["user"]; ok {
					idx.clients[conn].User = val
				}

				updateName := true
				if val, ok := body["uuid"]; ok {
					if val != "" {
						client.UUID = val
						updateName = false
					}
				}

				if updateName {
					idx.broadcast <- &Cmd{
						ClientUUID: client.UUID,
						Payload:    shared.NewMessage("setName", client.UUID, shared.EncodingText),
					}
				}
			}
		} else {
			Logger("clientMsgRouter", err.Error())
		}

		return true
	default:
		return false
	}

	return false
}
