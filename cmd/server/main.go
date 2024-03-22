package main

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type client struct {
	conn   *websocket.Conn
	room   string
	userID string
	send   chan []byte
}

var clients = make(map[string]map[*client]bool)
var broadcast = make(chan message)

type message struct {
	room   string
	userID string
	msg    []byte
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	userID := uuid.New().String()
	header := http.Header{}
	header.Set("Userid", userID)
	conn, _ := upgrader.Upgrade(w, r, header)
	room := r.URL.Query().Get("room")

	cl := &client{conn: conn, room: room, userID: userID, send: make(chan []byte)}
	if clients[room] == nil {
		clients[room] = make(map[*client]bool)
	}
	clients[room][cl] = true

	fmt.Printf("Client connected to room %s: %s\n", room, userID)

	go handleClientMessages(cl)

	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("Read message: %d, %s, %v\n", msgType, msg, err)
			break
		}
		broadcast <- message{room: room, userID: userID, msg: msg}
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
		fmt.Printf("Message received: %+v\n", msg)
		for client := range clients[msg.room] {
			if client.userID == msg.userID {
				continue
			}
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
