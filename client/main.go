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

var verbose *bool

func Debug(message string) {
	if *verbose == true {
		fmt.Println(message)
	}
}

func GetServer(server string) string {
	addr := url.URL{
		Scheme: "ws",
		Host:   fmt.Sprintf("%s", server),
		Path:   "/c",
	}

	return addr.String()
}

func main() {
	server := flag.String("server", "localhost:8000", "Server hostname.")
	verboseFlag := flag.Bool("verbose", false, "Additional debugging logs.")

	flag.Parse()

	verbose = verboseFlag
	shared.Logger = Debug

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	connection := shared.InitConnection()

	defer connection.Close()

	connection.ReadCallback = func(conn *shared.Conn, message *shared.Message) {
		fmt.Println(string(message.Serialize()))
		conn.Send(message)
	}

	go func() {
		for {
			Debug("Checking connection...")
			if connection.Establish(GetServer(*server)) {
				go connection.WritePump()
				go connection.ReadPump()
			}

			time.Sleep(30 * time.Second)
		}
	}()

	for {
		select {
		case <-interrupt:
			Debug("INTERRUPT!")
			connection.Close()
			os.Exit(0)
		}
	}
}
