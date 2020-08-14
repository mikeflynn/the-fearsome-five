package shared

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
	Logger  func(string, int)
)

type Conn struct {
	ws           *websocket.Conn
	SendChan     chan []byte
	ReadCallback func(*Conn, string)
	IsActive     bool
	IsReading    bool
	IsWriting    bool
	ParentID     int64
}

func (c *Conn) Write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

func (c *Conn) WritePump() {
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
		case message, ok := <-c.SendChan:
			if !ok {
				c.Write(websocket.CloseMessage, []byte{})
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
			Logger("Sending ping...", 1)
			if err := c.Write(websocket.PingMessage, nil); err != nil {
				Logger("PING ERROR: "+err.Error(), 1)
				return
			}
		}
	}

	c.IsWriting = false
}

func (c *Conn) ReadPump() {
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
				Logger(fmt.Sprintf("ERROR: %v", err), 1)

				c.Close()
			}

			break
		}

		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		Logger("Incoming: "+string(message), 1)

		if c.ReadCallback != nil {
			c.ReadCallback(c, string(message))
		}
	}

	c.IsReading = false
}

func (c *Conn) Establish(host string) bool {
	Logger("Connecting to "+host+"...", 1)

	ws, _, err := websocket.DefaultDialer.Dial(host, nil)
	if err == nil {
		Logger("Connection established!", 1)
		c.ws = ws
		c.IsActive = true

		c.ws.SetCloseHandler(func(code int, text string) error {
			Logger("Closing connection...", 1)
			c.IsActive = false
			return errors.New(text)
		})

		return true
	} else {
		Logger("Connection Error: "+err.Error(), 1)
	}

	return false
}

func (c *Conn) Close() {
	Logger("Closing connection...", 1)

	if c.IsActive {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			Logger(fmt.Sprintf("Caught panic: %v", r), 1)
			c.IsActive = false
		}
	}()

	err := c.ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		Logger("Write Close Error: "+err.Error(), 1)
	}

	if err := c.ws.Close(); err != nil {
		Logger("Error Closing Connection: "+err.Error(), 1)
	}

	c.IsActive = false
}

func (c *Conn) Send(message string) {
	c.SendChan <- []byte(message)
}
