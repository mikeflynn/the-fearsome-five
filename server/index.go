package main

import (
	"errors"

	"github.com/lithammer/shortuuid/v3"
	"github.com/mikeflynn/the-fearsome-five/shared"
)

type Index struct {
	clients    map[*shared.Conn]*Client
	register   chan *shared.Conn
	unregister chan *shared.Conn
	broadcast  chan *Cmd
	recieve    chan *Resp
}

type Client struct {
	UUID          string     `json:"uuid"`
	OS            string     `json:"os"`
	User          string     `json:"user"`
	waitingOnResp bool       `json:"-"`
	respChan      chan *Resp `json:"-"`
}

type Cmd struct {
	ClientUUID string
	Payload    *shared.Message
}

type Resp struct {
	ClientUUID string          `json:"from"`
	Payload    *shared.Message `json:"body"`
}

func initIndex() *Index {
	return &Index{
		broadcast:  make(chan *Cmd),
		recieve:    make(chan *Resp),
		register:   make(chan *shared.Conn),
		unregister: make(chan *shared.Conn),
		clients:    make(map[*shared.Conn]*Client),
	}
}

func getUUID() string {
	return shortuuid.New()
}

func (i *Index) start() {
	for {
		select {
		case conn := <-i.register:
			i.clients[conn] = &Client{
				UUID:          getUUID(),
				waitingOnResp: false,
				respChan:      make(chan *Resp),
			}
		case conn := <-i.unregister:
			if client, ok := i.clients[conn]; ok {
				delete(i.clients, conn)
				close(conn.SendChan)
				close(client.respChan)
			}
		case command := <-i.broadcast:
			conn, err := i.connByUUID(command.ClientUUID)
			if err == nil {
				conn.Send(command.Payload)
			}
		case response := <-i.recieve:
			client, err := i.clientByUUID(response.ClientUUID)
			if err != nil {
				Logger("SERVER", err.Error())
			}

			if client.waitingOnResp {
				client.respChan <- response
			}
		}
	}
}

func (i *Index) connByUUID(lookup string) (*shared.Conn, error) {
	for k, v := range i.clients {
		if lookup == v.UUID {
			return k, nil
		}
	}

	return &shared.Conn{}, errors.New("Client UUID not found.")
}

func (i *Index) clientByUUID(lookup string) (*Client, error) {
	conn, err := i.connByUUID(lookup)
	if err != nil {
		return &Client{}, err
	}

	return i.clients[conn], nil
}

func (i *Index) list() []*Client {
	ret := []*Client{}

	for _, v := range i.clients {
		ret = append(ret, v)
	}

	return ret
}
