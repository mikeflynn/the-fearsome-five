package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/auth0/go-jwt-middleware"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/websocket"
	"github.com/mikeflynn/the-fearsome-five/shared"
)

var Verbose bool = false
var StartTime time.Time
var Version string = "0.1"

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

		if !clientMsgRouter(idx, conn, message) {
			idx.recieve <- &Resp{
				ClientUUID: client.UUID,
				Payload:    message,
			}
		}
	}

	conn.CloseCallback = func(conn *shared.Conn) {
		idx.unregister <- conn
	}

	go conn.WritePump()
	go conn.ReadPump()
}

func adminRouter(idx *Index, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		switch r.URL.Path {
		case "/a/list":
			adminGetList(idx, w, r)
			return
		case "/a/status":
			adminGetStatus(idx, w, r)
			return
		default:
			http.Error(w, "Not found", http.StatusNotImplemented)
			return
		}
	case "POST":
		switch r.URL.Path {
		case "/a/run":
			adminRunCommand(idx, w, r)
			return
		case "/a/fileSend":
			adminFileSend(idx, w, r)
			return
		case "/a/fileReq":
			adminFileReq(idx, w, r)
			return
		default:
			http.Error(w, "Not found", http.StatusNotImplemented)
			return
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func Logger(prefix string, message string) {
	if Verbose {
		log.Println(fmt.Sprintf("%s > %s", prefix, message))
	}
}

func getAbsPath(path string) string {
	if !filepath.IsAbs(path) {
		if wd, err := os.Getwd(); err == nil {
			path = wd + path
		}
	}

	return path
}

func verifySSLFiles(key string, cert string) bool {
	if key == "" || cert == "" {
		return false
	}

	if _, err := os.Stat(getAbsPath(key)); err != nil {
		Logger("init", err.Error())
		return false
	}

	if _, err := os.Stat(getAbsPath(cert)); err != nil {
		Logger("init", err.Error())
		return false
	}

	return true
}

func main() {
	addr := flag.String("listen", "0.0.0.0:8000", "API listen address.")
	verbose := flag.Bool("verbose", false, "Display extra logging.")
	sslCert := flag.String("ssl-cert", "", "Cert file for SSL")
	sslKey := flag.String("ssl-key", "", "Key file for SSL")
	passCode := flag.String("password", "", "Key file for SSL")
	flag.Parse()

	Verbose = *verbose

	Logger("INIT", "Starting with verbose on!")

	shared.MaxMessageSize = 1024 * 1024 * 1024
	shared.Logger = func(message string) {
		Logger("LIBRARY", message)
	}

	StartTime = time.Now()

	index := initIndex()
	go index.start()

	adminHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//Logger("ADMIN REQ", fmt.Sprintf("%s:%s", r.Method, r.URL.Path))
		adminRouter(index, w, r)
	})

	if *passCode != "" {
		jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
			ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
				return []byte(*passCode), nil
			},
			SigningMethod: jwt.SigningMethodHS256,
		})

		http.Handle("/a/", jwtMiddleware.Handler(adminHandler))
	} else {
		http.Handle("/a/", adminHandler)
	}

	http.HandleFunc("/c", func(w http.ResponseWriter, r *http.Request) {
		clientRouter(index, w, r)
	})

	if verifySSLFiles(*sslKey, *sslCert) {
		Logger("INIT", "Starting secure server.")
		err := http.ListenAndServeTLS(*addr, *sslCert, *sslKey, nil)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	} else {
		Logger("INIT", "WARNING: Starting server insecurely!!!")
		err := http.ListenAndServe(*addr, nil)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}
}
