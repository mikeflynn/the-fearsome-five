package main

import (
	"bytes"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 60 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1024
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Conn struct {
	ws           *websocket.Conn
	send         chan []byte
	readCallback func(*Conn, string)
	IsActive     bool
}

func (c *Conn) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

func (c *Conn) writePump() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		c.ws.Close()
		ticker.Stop()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}

			c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			w, err := c.ws.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			w.Write(message)
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			REPLLog("Sending ping...", 1)
			if err := c.write(websocket.PingMessage, nil); err != nil {
				REPLLog("PING ERROR: "+err.Error(), 1)
				return
			}
		}
	}
}

func (c *Conn) readPump() {
	defer func() {
		c.ws.Close()
	}()

	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				REPLLog(fmt.Sprintf("ERROR: %v", err), 1)

				c.Close()
			}
			break
		}

		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		REPLLog("Incoming: "+string(message), 1)

		if c.readCallback != nil {
			c.readCallback(c, string(message))
		}
	}
}

func (c *Conn) Close() {
	REPLLog("Closing connection...", 1)

	err := c.ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		REPLLog("Write Close Error: "+err.Error(), 1)
	}

	if err := c.ws.Close(); err != nil {
		REPLLog("Error Closing Connection: "+err.Error(), 1)
	}

	c.IsActive = false
}
