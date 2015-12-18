package main

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

const (
	writeWait = 10 * time.Second

	pongWait = 60 * time.Second

	pingPeriod = (pongWait * 9) / 10

	maxMessageSize = 512
)

type Client struct {
	ws   *websocket.Conn
	data chan []byte
}

func (c *Client) read(r *Room) {
	defer func() {
		r.unregister <- c
		c.ws.Close()
	}()

	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(
		func(string) error {
			c.ws.SetReadDeadline(time.Now().Add(pongWait))
			return nil
		},
	)

	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
		r.broadcast <- message
	}
}

func (c *Client) write(r *Room) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()

	for {
		select {
		case message, ok := <-c.data:
			if !ok {
				c.set(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.set(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.set(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func (c *Client) set(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

func wsPokimon(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["room"]
	if room, ok := hub.rooms[key]; ok {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		c := &Client{data: make(chan []byte), ws: ws}
		room.register <- c
		go c.write(room)
		c.read(room)
	}
}
