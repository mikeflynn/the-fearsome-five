package main

import (
	//"bytes"
	"flag"
	"fmt"
	"os"
	"os/signal"
	//"strings"
	"time"

	"github.com/gorilla/websocket"
)

var server = flag.String("server", "localhost", "Server hostname.")
var port = flag.String("port", "8080", "Server port.")
var verbose = flag.Bool("verbose", false, "Additional debuggin logs.")

var CmdChan = make(chan string)
var Connection *websocket.Conn

func main() {
	flag.Parse()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	connect()
	defer Connection.Close()

	go func() {
		for {
			_, message, err := Connection.ReadMessage()
			if err != nil {
				debug("Read Error: " + err.Error())

				connect()
				defer Connection.Close()
			}
			debug(fmt.Sprintf("Incoming: %s\n", message))
		}
	}()

	for {
		select {
		case <-interrupt:
			debug("INTERRUPT!")

			err := Connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				debug("Write Close Error: " + err.Error())
				return
			}

			return
		}
	}
}

func send(conn *websocket.Conn, message string) error {
	return conn.WriteMessage(websocket.TextMessage, []byte(message))
}

func debug(message string) {
	if *verbose == true {
		fmt.Println(message)
	}
}

func getServer() string {
	return "ws://" + *server + ":" + *port + "/ready"
}

func connect() {
	host := getServer()

	debug("Conncting to " + host + "...")

	for {
		c, _, err := websocket.DefaultDialer.Dial(host, nil)
		if err == nil {
			debug("Connection established!")
			Connection = c
			return
		}

		if *verbose == true {
			debug("Connection Error: " + err.Error())
		}

		time.Sleep(5 * time.Second)
	}
}
