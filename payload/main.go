package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"
)

var server = flag.String("server", "localhost", "Server hostname.")
var port = flag.String("port", "8080", "Server port.")
var verbose = flag.Bool("verbose", false, "Additional debuggin logs.")

var Connection = &Conn{send: make(chan []byte, 256), IsActive: false}

func main() {
	flag.Parse()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	defer Connection.ws.Close()

	Connection.Establish()

	go func() {
		for {
			Debug("Checking connection...")
			if !Connection.IsActive {
				Connection.Establish()
			}

			time.Sleep(5 * time.Minute)
		}
	}()

	go Connection.writePump()

	Connection.readCallback = func(conn *Conn, message string) {
		//client.Connection.send <- []byte("Active")

		//conn.readCallback = nil
	}

	Connection.readPump()

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
