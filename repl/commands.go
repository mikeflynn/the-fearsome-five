package main

import (
	"fmt"

	"gopkg.in/abiosoft/ishell.v2"
)

func cmdDefault(sh *ishell.Context) {
	sh.Println("Command not yet implemented.")
}

func cmdPing(sh *ishell.Context) {
	for _, c := range ClientIndex.Clients {
		c.Send("PING")
	}
}

func cmdList(sh *ishell.Context) {
	message := ""

	for _, c := range ClientIndex.list() {
		message = message + fmt.Sprintf(`
====================================
Client: %v
Active: %v
`, c.ID, c.Connection.IsActive)
	}

	sh.Println(message)
}
