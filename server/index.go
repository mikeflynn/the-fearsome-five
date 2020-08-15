package main

import "github.com/mikeflynn/the-fearsome-five/shared"

type Index struct {
	clients    map[*shared.Conn]bool
	register   chan *shared.Conn
	unregister chan *shared.Conn
	broadcast  chan *Cmd
}

type Cmd struct {
	client  *shared.Conn
	payload string
}

func initIndex() *Index {
	return &Index{
		broadcast:  make(chan *Cmd),
		register:   make(chan *shared.Conn),
		unregister: make(chan *shared.Conn),
		clients:    make(map[*shared.Conn]bool),
	}
}

func (i *Index) start() {
	for {
		select {
		case client := <-i.register:
			i.clients[client] = true
		case client := <-i.unregister:
			if _, ok := i.clients[client]; ok {
				delete(i.clients, client)
				close(client.SendChan)
			}
		case command := <-i.broadcast:
			if _, ok := i.clients[command.client]; ok {
				command.client.Send(command.payload)
			}
		}
	}
}
