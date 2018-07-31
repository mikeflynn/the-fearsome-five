package main

import (
	"gopkg.in/abiosoft/ishell.v2"
)

func cmdDefault(c *ishell.Context) {
	c.Println("Command not yet implemented.")
}
func cmdPing(c *ishell.Context) {
	for _, c := range ClientIndex.Clients {
		c.Send("PING")
	}
}
