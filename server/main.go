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

func adminRouter(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not found", http.StatusNotFound)
	return
}

func clientRouter(idx *Index, w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	ws, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		Logger("SERVER", fmt.Sprintf("Connection Issue: %s", err.Error()))
		return
	}

	conn := shared.InitConnection()
	conn.SetWS(ws) // Since this is the server, set the web socket without Establish()

	idx.register <- conn

	Logger("SERVER", "Client connected!")
	Logger("SERVER", fmt.Sprintf("Total Clients: %d", len(idx.clients)))

	conn.ReadCallback = func(conn *shared.Conn, message string) {
		Logger("FROM CLIENT", message)
	}

	conn.CloseCallback = func(conn *shared.Conn) {
		idx.unregister <- conn
	}

	go conn.WritePump()
	go conn.ReadPump()
}

func Logger(prefix string, message string) {
	if *Verbose {
		log.Println(fmt.Sprintf("%s > %s", prefix, message))
	}
}

func main() {
	addr := flag.String("listen", "0.0.0.0:8000", "API listen address.")
	Verbose = flag.Bool("verbose", false, "Display extra logging.")

	flag.Parse()

	shared.Logger = func(message string) {
		Logger("LIBRARY", message)
	}

	index := initIndex()
	go index.start()

	http.HandleFunc("/a", adminRouter)
	http.HandleFunc("/c", func(w http.ResponseWriter, r *http.Request) {
		clientRouter(index, w, r)
	})

	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
