package main

import (
	"fmt"

	"github.com/fatih/color"
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
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	for _, c := range ClientIndex.list() {
		active := red("✖︎")
		if c.Connection.IsActive == true {
			active = green("✔︎")
		}

		message = message + fmt.Sprintf("%v [%v] %s @ %s\n", c.ID, active, c.Username, c.OperatingSystem)
	}

	sh.Println(message)
}
