package main

// import (
// 	"log"
// 	"time"
// )

type Room struct {
	name       string
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	over       chan bool
}

func (r *Room) handle() {
	for {
		select {
		case c := <-r.register:
			r.clients[c] = true
		case c := <-r.unregister:
			if _, ok := r.clients[c]; ok {
				delete(r.clients, c)
				close(c.data)
			}
		case m := <-r.broadcast:
			for c := range r.clients {
				select {
				case c.data <- m:
				default:
					close(c.data)
					delete(r.clients, c)
				}
			}
		case <-r.over:
			return
		}
	}

	// for {
	// 	select {
	// 	case <-r.over:
	// 		return
	// 	default:
	// 		log.Println("room handle: ", r)
	// 		time.Sleep(5 * time.Second)
	// 	}
	// }
}
