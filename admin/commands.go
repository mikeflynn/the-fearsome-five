package main

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"gopkg.in/abiosoft/ishell.v2"
)

var currentFilters url.Values = url.Values{}

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
		Func: cmdSetFilter,
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "clear_filter",
		Help: "Resets all client filters.",
		Func: cmdClearFilter,
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
	list := []*Client{}
	filters := currentFilters.Encode()

	if filters == "" {
		list = ClientIndex.list()
	} else {
		list, _ = ClientIndex.filter(filters)
	}

	message := ""
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	for _, c := range list {
		active := red("✖︎")
		if c.Connection.IsActive == true {
			active = green("✔︎")
		}

		message = message + fmt.Sprintf("%v [%v] %s@%s (%s)\n", c.ID, active, c.Username, c.ExternalIP, c.OperatingSystem)
	}

	sh.Println(message)
}

func cmdDetail(sh *ishell.Context) {
	if len(sh.Args) == 0 {
		sh.Println("Missing argument.")
		return
	}

	id, err := strconv.ParseInt(sh.Args[0], 10, 64)
	if err != nil {
		sh.Println(err.Error())
		return
	}

	row, err := ClientIndex.getClientByID(id)
	if err != nil {
		sh.Println(err.Error())
		return
	}

	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	active := red("✖︎")
	if row.Connection.IsActive == true {
		active = green("✔︎")
	}

	sh.Println(fmt.Sprintf(`
Client Detail:
ID: %v [%v]
User: %s@%s
OS: %v
Groups: %v
Tags: %v`, row.ID, active, row.Username, row.ExternalIP, row.OperatingSystem, row.Groups, row.Tags))

}

func cmdSetFilter(sh *ishell.Context) {
	if len(sh.Args) == 0 {
		filter := currentFilters.Encode()
		if filter == "" {
			sh.Println("No filters are set.")
		} else {
			sh.Println("Current Filter: " + currentFilters.Encode())
		}
	} else {
		parts := strings.Split(sh.Args[0], "=")
		currentFilters.Set(parts[0], parts[1])
		sh.Println("Filter updated.")
	}
}

func cmdClearFilter(sh *ishell.Context) {
	currentFilters = url.Values{}
}
