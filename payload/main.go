package main

import (
	//"bytes"
	"flag"
	"fmt"
	"os"
	"os/signal"
	//"strings"
	//"time"

	"github.com/gorilla/websocket"
)

var server = flag.String("server", "localhost", "Server hostname.")
var port = flag.String("port", "8080", "Server port.")
var verbose = flag.Bool("verbose", false, "Additional debuggin logs.")

var CmdChan = make(chan string)

func main() {
	flag.Parse()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	host := getServer()
	debug("Conncting to " + host + "...")

	c, _, err := websocket.DefaultDialer.Dial(host, nil)
	if err != nil {
		debug("Dial Error: " + err.Error())
	}

	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				debug("Read Error: " + err.Error())
				return
			}
			debug(fmt.Sprintf("Incoming: %s\n", message))
		}
	}()

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			debug("INTERRUPT!")

			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				debug("Write Close Error: " + err.Error())
				return
			}

			select {
			case <-done:
			}

			return
		}
	}
}

func debug(message string) {
	if *verbose == true {
		fmt.Println(message)
	}
}

func getServer() string {
	return "ws://" + *server + ":" + *port + "/ready"
}
