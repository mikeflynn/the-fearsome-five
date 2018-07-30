package main

import (
	"strings"

	"gopkg.in/abiosoft/ishell.v2"
)

func cmdDefault(c *ishell.Context) {
	c.Println("Command not yet implemented.")
}

func cmdGreet(c *ishell.Context) {
	c.Println("Hello", strings.Join(c.Args, " "))
}
