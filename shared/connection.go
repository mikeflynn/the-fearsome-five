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
	Logger  func(string)
)

type Conn struct {
	Ws            *websocket.Conn
	SendChan      chan []byte
	ReadCallback  func(*Conn, string)
	CloseCallback func(*Conn)
	State         int // 0 = ready; -1 = closed
}

func InitConnection() *Conn {
	return &Conn{
		SendChan: make(chan []byte, 256),
		State:    -1,
	}
}

func (c *Conn) Write(mt int, payload []byte) error {
	c.Ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.Ws.WriteMessage(mt, payload)
}

func (c *Conn) WritePump() {
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

			c.Ws.SetWriteDeadline(time.Now().Add(writeWait))
			w, err := c.Ws.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			w.Write(message)
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			Logger("Sending ping...")
			if err := c.Write(websocket.PingMessage, nil); err != nil {
				Logger("PING ERROR: " + err.Error())
				return
			}
		}
	}
}

func (c *Conn) ReadPump() {
	defer func() {
		c.Close()
	}()

	c.Ws.SetReadLimit(maxMessageSize)
	c.Ws.SetReadDeadline(time.Now().Add(pongWait))
	c.Ws.SetPongHandler(func(string) error { c.Ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.Ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				Logger(fmt.Sprintf("ERROR: %v", err))

				c.Close()
			}

			break
		}

		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		Logger("Incoming: " + string(message))

		if c.ReadCallback != nil {
			c.ReadCallback(c, string(message))
		}
	}
}

func (c *Conn) Establish(host string) bool {
	Logger("Connecting to " + host + "...")

	ws, _, err := websocket.DefaultDialer.Dial(host, nil)
	if err == nil {
		Logger("Connection established!")
		c.SetWS(ws)

		c.Ws.SetCloseHandler(func(code int, text string) error {
			Logger("Closing connection...")
			return errors.New(text)
		})

		return true
	} else {
		Logger("Connection Error: " + err.Error())
	}

	return false
}

func (c *Conn) SetWS(ws *websocket.Conn) {
	c.Ws = ws
	c.State = 0
}

func (c *Conn) Close() {
	if c.State == -1 {
		return
	}

	c.State = -1

	Logger("Closing connection...")

	defer func() {
		if r := recover(); r != nil {
			Logger(fmt.Sprintf("Caught panic: %v", r))
		}
	}()

	if c.CloseCallback != nil {
		c.CloseCallback(c)
	}

	err := c.Ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		Logger("Write Close Error: " + err.Error())
	}

	if err := c.Ws.Close(); err != nil {
		Logger("Error Closing Connection: " + err.Error())
	}
}

func (c *Conn) Send(message string) {
	c.SendChan <- []byte(message)
}
