package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", ":8080", "http service address")
var homeTempl = template.Must(template.ParseFiles("index.html"))
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	fmt.Println("server start....")
	go hub.run()
	flag.Parse()
	r := mux.NewRouter()
	r.HandleFunc("/", serveHome)
	r.HandleFunc("/ws", serveWs)
	r.HandleFunc("/pokimon/new/{room:[a-zA-Z0-9]+}", newRoom)
	r.HandleFunc("/pokimon/delete/{room:[a-zA-Z0-9]+}", delRoom)
	r.HandleFunc("/ws/pokimon/{room:[a-zA-Z0-9]+}", wsPokimon)

	err := http.ListenAndServe(*addr, r)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func newRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	vars := mux.Vars(r)
	key := vars["room"]

	log.Println("hub exec")
	go hub.newRoom(key)
	log.Println("hub ", hub)

	ret := make(map[string]interface{})
	ret["url"] = "ws://" + r.Host + "/ws/pokimon/" + key
	ret["timestamp"] = time.Now().UTC().Unix()

	w.Header().Set("Content-Type", "text/json; charset=utf-8")
	err := json.NewEncoder(w).Encode(ret)

	if err != nil {
		log.Println(r.RemoteAddr, " error: ", err)
		return
	}

	log.Println(r.RemoteAddr)
}

func delRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	vars := mux.Vars(r)
	key := vars["room"]

	ret := make(map[string]interface{})
	bools := hub.delRoom(key)
	ret["result"] = bools

	w.Header().Set("Content-Type", "text/json; charset=utf-8")
	err := json.NewEncoder(w).Encode(ret)

	if err != nil {
		log.Println(r.RemoteAddr, " error: ", err)
		return
	}

	log.Println(r.RemoteAddr)
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for t := range ticker.C {
			ws.WriteJSON(t)
		}
	}()
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	log.Println(r.Host)
	homeTempl.Execute(w, r.Host)
}
