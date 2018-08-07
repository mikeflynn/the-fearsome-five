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
			Debug("Sending ping...")
			if err := c.write(websocket.PingMessage, nil); err != nil {
				Debug("PING ERROR: " + err.Error())
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
				Debug("UNEXPECTED ERROR: " + err.Error())

				c.Close()
			}

			break
		}

		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		Debug("Incoming: " + string(message))

		if c.readCallback != nil {
			c.readCallback(c, string(message))
		}
	}

	c.IsReading = false
}

func (c *Conn) Establish(host string) bool {
	Debug("Connecting to " + host + "...")

	ws, _, err := websocket.DefaultDialer.Dial(host, nil)
	if err == nil {
		Debug("Connection established!")
		c.ws = ws
		c.IsActive = true

		c.ws.SetCloseHandler(func(code int, text string) error {
			Debug("Closing connection...")
			c.IsActive = false
			return errors.New(text)
		})

		return true
	} else {
		Debug("Connection Error: " + err.Error())
	}

	return false
}

func (c *Conn) Close() {
	Debug("Closing connection...")

	if c.IsActive {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			Debug(fmt.Sprintf("Caught panic: %v", r))
			c.IsActive = false
		}
	}()

	err := c.ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		Debug("Write Close Error: " + err.Error())
	}

	if err := c.ws.Close(); err != nil {
		Debug("Error Closing Connection: " + err.Error())
	}

	c.IsActive = false
}

func (c *Conn) Send(message string) {
	c.send <- []byte(message)
}
