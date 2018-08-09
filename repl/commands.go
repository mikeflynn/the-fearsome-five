package main

import (
	"fmt"

	"github.com/fatih/color"
	"gopkg.in/abiosoft/ishell.v2"
)

func InitCommands(shell *ishell.Shell) {

	shell.AddCmd(&ishell.Cmd{
		Name: "ping",
		Help: "Sends ping",
		Func: func(sh *ishell.Context) {
			for _, c := range ClientIndex.Clients {
				c.Send("PING")
			}
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "filter",
		Help: "Sets global filters for subsequent commands (ex. os=mac).",
		Func: cmdDefault,
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "list",
		Help: "List active clients.",
		Func: cmdList,
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "detail",
		Help: "Shows the full details on a given client.",
		Func: cmdDetail,
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "shell",
		Help: "Opens shell on a specific client.",
		Func: cmdDefault,
	})
}

func cmdDefault(sh *ishell.Context) {
	sh.Println("Command not yet implemented.")
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

		message = message + fmt.Sprintf("%v [%v] %s@%s (%s)\n", c.ID, active, c.Username, c.ExternalIP, c.OperatingSystem)
	}

	sh.Println(message)
}

func cmdDetail(sh *ishell.Context) {

}
