package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type client struct {
	conn *websocket.Conn
	room string
	send chan []byte
}

var clients = make(map[string]map[*client]bool)
var broadcast = make(chan message)

type message struct {
	room string
	msg  []byte
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, _ := upgrader.Upgrade(w, r, nil)
	room := r.URL.Query().Get("room")

	cl := &client{conn: conn, room: room, send: make(chan []byte)}
	if clients[room] == nil {
		clients[room] = make(map[*client]bool)
	}
	clients[room][cl] = true

	go handleClientMessages(cl)

	for {
		_, msg, _ := conn.ReadMessage()
		broadcast <- message{room: room, msg: msg}
	}
}

func handleClientMessages(client *client) {
	defer func() {
		client.conn.Close()
		delete(clients[client.room], client)
	}()

	for {
		select {
		case msg, ok := <-client.send:
			if !ok {
				return
			}
			client.conn.WriteMessage(websocket.TextMessage, msg)
		}
	}
}

func handleMessages() {
	for {
		msg := <-broadcast
		for client := range clients[msg.room] {
			select {
			case client.send <- msg.msg:
			default:
				close(client.send)
				delete(clients[msg.room], client)
			}
		}
	}
}

func main() {
	http.HandleFunc("/ws", handleConnections)
	go handleMessages()

	fmt.Println("Server started on :8080")
	http.ListenAndServe(":8080", nil)
}
