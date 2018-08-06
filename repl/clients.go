package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"math"
	"math/rand"
	"time"
)

var (
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

func (this *Client) Send(message string) {
	this.Connection.send <- []byte(message)
}
