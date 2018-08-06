package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"time"
)

var server = flag.String("server", "localhost", "Server hostname.")
var port = flag.String("port", "8080", "Server port.")
var verbose = flag.Bool("verbose", false, "Additional debuggin logs.")

var Connection = &Conn{send: make(chan []byte, 256), IsActive: false, IsReading: false, IsWriting: false}

func main() {
	flag.Parse()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	defer Connection.Close()

	Connection.readCallback = func(conn *Conn, message string) {
		//client.Connection.send <- []byte("Active")

		params, err := url.ParseQuery(message)
		if err != nil {
			Debug("Unable to parse message")
			return
		}

		switch mtype := params.Get("type"); mtype {
		case "init":
			Debug(fmt.Sprintf("Initialized with new client id: %v", params.Get("client_id")))
		default:
			Debug("Unknown message type.")
		}
	}

	go func() {
		for {
			Debug("Checking connection...")
			if !Connection.IsActive {
				if Connection.Establish() {
					go Connection.writePump()
					Connection.readPump()
				}
			}

			time.Sleep(5 * time.Second)
		}
	}()

	for {
		select {
		case <-interrupt:
			Debug("INTERRUPT!")
			Connection.Close()
			os.Exit(0)
		}
	}
}

func Debug(message string) {
	if *verbose == true {
		fmt.Println(message)
	}
}

func GetServer() string {
	return "ws://" + *server + ":" + *port + "/ready"
}
