package main

import (
	"bytes"
	"errors"
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
	logger  func(string, int)
)

type Conn struct {
	ws           *websocket.Conn
	send         chan []byte
	readCallback func(*Conn, string)
	IsActive     bool
	IsReading    bool
	IsWriting    bool
	ParentID     int64
}

func (c *Conn) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

func (c *Conn) writePump() {
	if c.IsWriting == true {
		return
	}

	c.IsWriting = true

	ticker := time.NewTicker(pingPeriod)

	defer func() {
		c.Close()
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
			logger("Sending ping...", 1)
			if err := c.write(websocket.PingMessage, nil); err != nil {
				logger("PING ERROR: "+err.Error(), 1)
				return
			}
		}
	}

	c.IsWriting = false
}

func (c *Conn) readPump() {
	if c.IsReading == true {
		return
	}

	c.IsReading = true

	defer func() {
		c.Close()
	}()

	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				logger("UNEXPECTED ERROR: "+err.Error(), 1)

				c.Close()
			}

			break
		}

		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		logger("Incoming: "+string(message), 1)

		if c.readCallback != nil {
			c.readCallback(c, string(message))
		}
	}

	c.IsReading = false
}

func (c *Conn) Establish(host string) bool {
	logger("Connecting to "+host+"...", 1)

	ws, _, err := websocket.DefaultDialer.Dial(host, nil)
	if err == nil {
		logger("Connection established!", 1)
		c.ws = ws
		c.IsActive = true

		c.ws.SetCloseHandler(func(code int, text string) error {
			logger("Closing connection...", 1)
			c.IsActive = false
			return errors.New(text)
		})

		return true
	} else {
		logger("Connection Error: "+err.Error(), 1)
	}

	return false
}

func (c *Conn) Close() {
	logger("Closing connection...", 1)

	if c.IsActive {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			logger(fmt.Sprintf("Caught panic: %v", r), 1)
			c.IsActive = false
		}
	}()

	err := c.ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		logger("Write Close Error: "+err.Error(), 1)
	}

	if err := c.ws.Close(); err != nil {
		logger("Error Closing Connection: "+err.Error(), 1)
	}

	c.IsActive = false
}

func (c *Conn) Send(message string) {
	c.send <- []byte(message)
}
