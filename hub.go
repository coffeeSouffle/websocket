package main

import (
	"log"

	// "github.com/gorilla/mux"
	// "github.com/gorilla/websocket"
)

var hub = Hub{
	rooms:    make(map[string]*Room),
	roomChan: make(chan *Room),
	del:      make(chan *Room),
}

type Hub struct {
	rooms    map[string]*Room
	roomChan chan *Room
	del      chan *Room
}

func (h *Hub) newRoom(key string) {
	var room *Room
	log.Println("exec hub new Room")
	if _, ok := h.rooms[key]; ok {
		return
	}

	room = &Room{
		name:       key,
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		over:       make(chan bool),
	}
	log.Println("new room", room)
	h.rooms[key] = room
	h.roomChan <- room
}

func (h *Hub) delRoom(key string) bool {
	if room, ok := h.rooms[key]; ok {
		if len(room.clients) == 0 {
			room.over <- true
			h.del <- room
			return true
		}
	}

	return false
}

func (h *Hub) run() {
	for {
		select {
		case r := <-h.roomChan:
			log.Println("hub run: ", r)
			go r.handle()
		case r := <-h.del:
			log.Println("hub delete room:", r)
			for k, v := range h.rooms {
				if r.name == v.name {
					delete(h.rooms, k)
				}
			}
		}
	}
}
