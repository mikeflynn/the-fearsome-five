package main

import (
	"encoding/json"

	"github.com/mikeflynn/the-fearsome-five/shared"
)

func clientMsgRouter(idx *Index, conn *shared.Conn, message *shared.Message) bool {
	switch message.Action {
	case "systemReport":
		body, err := message.GetBodyJSON()

		if err == nil {
			if client, yes := idx.clients[conn]; yes {
				jsonBody, err := json.Marshal(body)
				if err != nil {
					Logger("systemReport", err.Error())
					return false
				}

				if err := json.Unmarshal(jsonBody, &client.Report); err != nil {
					Logger("systemReport", err.Error())
					return false
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
