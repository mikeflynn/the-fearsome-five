package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/mikeflynn/the-fearsome-five/shared"
)

var Verbose bool = false

func adminRouter(idx *Index, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		switch r.URL.Path {
		case "/a/list":
			resp := idx.list()

			if err := json.NewEncoder(w).Encode(resp); err != nil {
				http.Error(w, "Internal error", http.StatusInternalServerError)
			}
		default:
			http.Error(w, "Not found", http.StatusNotImplemented)
			return
		}
	case "POST":
		switch r.URL.Path {
		case "/a/send":
			cid, ok := r.URL.Query()["cid"]
			if !ok || len(cid[0]) < 1 {
				http.Error(w, "Internal error", http.StatusBadRequest)
				return
			}

			msg, ok := r.URL.Query()["msg"]
			if !ok || len(msg[0]) < 1 {
				http.Error(w, "Internal error", http.StatusBadRequest)
				return
			}

			doWait := false
			waitParam, ok := r.URL.Query()["wait"]
			if ok && len(waitParam[0]) == 1 {
				doWait = true
			}

			client, err := idx.clientByUUID(cid[0])
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			cmd := &Cmd{
				ClientUUID: client.UUID,
				Payload:    shared.NewMessage("std", msg[0]),
			}

			if doWait {
				client.waitingOnResp = true
			}

			idx.broadcast <- cmd

			if doWait {
				resp := <-client.respChan
				client.waitingOnResp = false

				if err := json.NewEncoder(w).Encode(resp); err != nil {
					http.Error(w, "Internal error", http.StatusInternalServerError)
				}
			} else {
				if err := json.NewEncoder(w).Encode(map[string]bool{"success": true}); err != nil {
					http.Error(w, "Internal error", http.StatusInternalServerError)
				}
			}
		default:
			http.Error(w, "Not found", http.StatusNotImplemented)
			return
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
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

	conn.ReadCallback = func(conn *shared.Conn, message *shared.Message) {
		client := idx.clients[conn]
		Logger(fmt.Sprintf("%s:", client.UUID), string(message.Serialize()))

		idx.recieve <- &Resp{
			ClientUUID: client.UUID,
			Payload:    message,
		}
	}

	conn.CloseCallback = func(conn *shared.Conn) {
		idx.unregister <- conn
	}

	go conn.WritePump()
	go conn.ReadPump()
}

func Logger(prefix string, message string) {
	if Verbose {
		log.Println(fmt.Sprintf("%s > %s", prefix, message))
	}
}

func main() {
	addr := flag.String("listen", "0.0.0.0:8000", "API listen address.")
	verbose := flag.Bool("verbose", false, "Display extra logging.")
	flag.Parse()

	Verbose = *verbose

	Logger("INIT", "Starting with verbose on!")

	shared.Logger = func(message string) {
		Logger("LIBRARY", message)
	}

	index := initIndex()
	go index.start()

	http.HandleFunc("/a/", func(w http.ResponseWriter, r *http.Request) {
		Logger("ADMIN REQ", fmt.Sprintf("%s:%s", r.Method, r.URL.Path))
		adminRouter(index, w, r)
	})

	http.HandleFunc("/c", func(w http.ResponseWriter, r *http.Request) {
		clientRouter(index, w, r)
	})

	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
