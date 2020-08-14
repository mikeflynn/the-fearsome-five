package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/mikeflynn/the-fearsome-five/shared"
)

var Verbose *bool

func main() {
	clientPort := flag.String("client-port", "8000", "Client connection port.")
	adminPort := flag.String("admin-port", "9000", "Admin connection port.")
	Verbose := flag.Bool("verbose", false, "Display extra logging.")

	startServer(*clientPort)
}

func clientRouter(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	ws, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		Logger("SERVER", fmt.Sprintf("Connection Issue: %s", err.Error()))
		return
	}

	conn := &shared.Conn{SendChan: make(chan []byte, 256), Ws: ws, IsActive: true}
	go conn.WritePump()

	conn.ReadCallback = func(conn *shared.Conn, message string) {
		Logger("FROM CLIENT", message)
	}

	conn.ReadPump()
}

func startServer(port string) {
	http.HandleFunc("/ready", clientRouter)
	http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", port), nil)
}

func Logger(prefix string, message string) {
	if *Verbose {
		log.Println(fmt.Sprintf("%s > %s"), prefix, message)
	}
}
