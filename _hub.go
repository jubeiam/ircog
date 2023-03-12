package main

import (
	"fmt"
	"log"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool
	rooms   map[*Client][]string

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	join chan struct {
		c    *Client
		room string
	}

	part chan struct {
		c    *Client
		room string
	}

	privmsg chan struct {
		msg  []byte
		room string
	}

	// Unregister requests from clients.
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		join: make(chan struct {
			c    *Client
			room string
		}),
		part: make(chan struct {
			c    *Client
			room string
		}),
		privmsg: make(chan struct {
			msg  []byte
			room string
		}),
		clients: make(map[*Client]bool),
		rooms:   make(map[*Client][]string),
	}
}

func filterClients(s []*Client, c *Client) []*Client {
	var out []*Client

	for _, client := range s {
		if client != c {
			out = append(out, client)
		}
	}

	return out
}

func filterStrings(in []string, match string) []string {
	var out []string

	for _, s := range in {
		if match != s {
			out = append(out, s)
		}
	}

	return out
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				delete(h.rooms, client)
				close(client.send)
			}
		case data := <-h.join:
			h.rooms[data.c] = append(h.rooms[data.c], data.room)
			data.c.send <- []byte(fmt.Sprintf("JOIN %s", data.room))

			log.Printf("%d in room %s", len(h.rooms[data.c]), data.room)
		case data := <-h.part:
			h.rooms[data.c] = filterStrings(h.rooms[data.c], data.room)
			data.c.send <- []byte(fmt.Sprintf("PART %s", data.room))

			log.Printf("%d in room %s", len(h.rooms[data.c]), data.room)
		// case message := <-h.privmsg:
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					delete(h.clients, client)
					close(client.send)
				}
			}
		}
	}
}
