package main

import (
	"bufio"
	"fmt"
	"net/url"
	"os"

	"github.com/gorilla/websocket"
)

func main() {
	fmt.Println("WebSocket Chat Client")
	fmt.Print("Enter room name: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	room := scanner.Text()
	serverURL := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws", RawQuery: "room=" + room}

	conn, resp, err := websocket.DefaultDialer.Dial(serverURL.String(), nil)
	if err != nil {
		fmt.Println("Error connecting to WebSocket server:", err)
		return
	}
	defer conn.Close()
	fmt.Println(resp.Header.Get("Userid"))

	go receiveMessages(conn)

	for scanner.Scan() {
		msg := scanner.Text()
		err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			fmt.Println("Error sending message:", err)
			break
		}
	}
}

func receiveMessages(conn *websocket.Conn) {
	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("Error receiving message: message type: %d, message: %s, error: %v\n", msgType, string(msg), err)
			return
		}
		fmt.Printf("Received message: %s\n", string(msg))
	}
}
