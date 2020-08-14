package main

import (
	"flag"
	"fmt"
	//"math/rand"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/mikeflynn/the-fearsome-five/shared"
)

var Connection = &shared.Conn{SendChan: make(chan []byte, 256), IsActive: false, IsReading: false, IsWriting: false}
var verbose *bool

func main() {
	server := flag.String("server", "localhost:8000", "Server hostname.")
	verboseFlag := flag.Bool("verbose", false, "Additional debugging logs.")

	flag.Parse()

	verbose = verboseFlag
	shared.Logger = Debug

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	defer Connection.Close()

	Connection.ReadCallback = func(conn *shared.Conn, message string) {
		IncomingRouter(conn, message)
	}

	go func() {
		for {
			Debug("Checking connection...")
			if !Connection.IsActive {
				if Connection.Establish(GetServer(*server)) {
					go Connection.WritePump()
					Connection.ReadPump()
				}
			}

			time.Sleep(30 * time.Second)
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

func GetServer(server string) string {
	addr := url.URL{
		Scheme: "ws",
		Host:   fmt.Sprintf("%s", server),
		Path:   "/ready",
	}

	return addr.String()
}
