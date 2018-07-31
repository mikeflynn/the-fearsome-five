package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	newline     = []byte{'\n'}
	space       = []byte{' '}
	connIdx     = map[string]*Conn{}
	ClientIndex = &Index{}
)

func genClientID() int64 {
	now := time.Now().Unix()
	rand.Seed(now)
	return int64(math.Ceil(float64(time.Now().Unix() * rand.Int63n(999999) / 100000000)))
}

// Index

type Index struct {
	Clients []*Client `json:"clients"`
}

func (this *Index) fileLoad(filename string) error {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(file, this); err != nil {
		return err
	}

	return nil
}

func (this *Index) fileSave(filename string) error {
	data, err := json.Marshal(this)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		return err
	}

	return nil
}

func (this *Index) addClient(ID int64, conn *Conn) *Client {
	if ID == 0 {
		ID = genClientID()
	}

	client := &Client{
		ID:         ID,
		Connection: conn,
	}

	this.Clients = append(this.Clients, client)

	return client
}

func (this *Index) getClientByID(id int64) (*Client, error) {
	for _, client := range this.Clients {
		if client.ID == id {
			return client, nil
		}
	}

	return &Client{}, errors.New("Client not found.")
}

// Client

type Client struct {
	ID              int64    `json:"id"`
	OperatingSystem string   `json:"operating_system"`
	Bandwidth       string   `json:"bandwidth"`
	Admin           bool     `json:"admin"`
	HasCamera       bool     `json:"has_camera"`
	HasMic          bool     `json:"has_mic"`
	Tags            []string `json:"tags"`
	LastConnection  int64    `json:"last_connection"`
	Connection      *Conn
}

func (this *Client) isActive() bool {
	timeout, _ := time.ParseDuration("-5m")
	if this.LastConnection > time.Now().Add(timeout).Unix() {
		return true
	}

	return false
}

// Connection

type Conn struct {
	ws           *websocket.Conn
	send         chan []byte
	readCallback func(*Conn, string)
}

func (c *Conn) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

func (c *Conn) writePump() {
	defer c.ws.Close()

	for {
		message, ok := <-c.send
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
				REPLLog(fmt.Sprintf("ERROR: %v", err))
			}
			break
		}

		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		REPLLog(string(message))

		if c.readCallback != nil {
			c.readCallback(c, string(message))
		}
	}
}
