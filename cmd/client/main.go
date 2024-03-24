package main

import (
	"bufio"
	"context"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"syscall"

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

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	go func() {
		receiveMessages(ctx, conn)
	}()

	for scanner.Scan() {
		msg := scanner.Text()
		err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			fmt.Println("Error sending message:", err)
			break
		}
	}

	fmt.Println(scanner.Err())
	fmt.Println("Client shutting down successfully")
}

func receiveMessages(ctx context.Context, conn *websocket.Conn) {
	defer fmt.Println("Stopped receiving messages")
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				fmt.Printf("Error receiving message: message type: %d, message: %s, error: %v\n", msgType, string(msg), err)
				return
			}
			fmt.Printf("Received message: %s\n", string(msg))
		}
	}
}
