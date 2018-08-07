package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	//"math/rand"
	"net/url"
	"os"
	"os/signal"
	"time"
)

var server = flag.String("server", "localhost", "Server hostname.")
var port = flag.String("port", "8080", "Server port.")
var verbose = flag.Bool("verbose", false, "Additional debuggin logs.")
var cid = flag.String("cid", "", "Override CID")

var Connection = &Conn{send: make(chan []byte, 256), IsActive: false, IsReading: false, IsWriting: false}

func main() {
	flag.Parse()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	defer Connection.Close()

	Connection.readCallback = func(conn *Conn, message string) {
		IncomingRouter(conn, message)
	}

	go func() {
		for {
			Debug("Checking connection...")
			if !Connection.IsActive {
				if Connection.Establish(GetServer()) {
					go Connection.writePump()
					Connection.readPump()
				}
			}

			time.Sleep(5 * time.Minute)
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
	addr := url.URL{
		Scheme: "ws",
		Host:   fmt.Sprintf("%s:%s", *server, *port),
		Path:   "/ready",
	}

	if *cid != "" {
		addr.RawQuery = fmt.Sprintf("cid=%s", *cid)
	} else {
		data, err := ioutil.ReadFile(GetConfigPath())
		if err == nil && string(data) != "" {
			addr.RawQuery = fmt.Sprintf("cid=%s", string(data))
		}
	}

	return addr.String()
}
