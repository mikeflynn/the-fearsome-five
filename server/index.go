package main

import (
	"errors"

	"github.com/lithammer/shortuuid/v3"
	"github.com/mikeflynn/the-fearsome-five/shared"
)

type Index struct {
	clients    map[*shared.Conn]string
	register   chan *shared.Conn
	unregister chan *shared.Conn
	broadcast  chan *Cmd
}

type Cmd struct {
	clientUUID string
	payload    *shared.Message
}

func initIndex() *Index {
	return &Index{
		broadcast:  make(chan *Cmd),
		register:   make(chan *shared.Conn),
		unregister: make(chan *shared.Conn),
		clients:    make(map[*shared.Conn]string),
	}
}

func getUUID() string {
	return shortuuid.New()
}

func (i *Index) start() {
	for {
		select {
		case client := <-i.register:
			i.clients[client] = getUUID()
		case client := <-i.unregister:
			if _, ok := i.clients[client]; ok {
				delete(i.clients, client)
				close(client.SendChan)
			}
		case command := <-i.broadcast:
			conn, err := i.clientByUUID(command.clientUUID)
			if err == nil {
				conn.Send(command.payload)
			}
		}
	}
}

func (i *Index) clientByUUID(lookup string) (*shared.Conn, error) {
	for k, v := range i.clients {
		if lookup == v {
			return k, nil
		}
	}

	return &shared.Conn{}, errors.New("Client UUID not found.")
}

func (i *Index) list() map[string]map[string]string {
	ret := map[string]map[string]string{}

	for _, v := range i.clients {
		ret[v] = map[string]string{
			"foo": "bar",
		}
	}

	return ret
}
