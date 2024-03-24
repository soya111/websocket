package main

import (
	"bufio"
	"context"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"sync"
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

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		receiveMessages(ctx, conn)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		sendMessages(ctx, conn)
	}()

	wg.Wait()
	fmt.Println("Client shutting down successfully")
}

func sendMessages(ctx context.Context, conn *websocket.Conn) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if !scanner.Scan() {
				fmt.Printf("Error reading from stdin: %v\n", scanner.Err())
				return
			}
			msg := scanner.Text()
			err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				fmt.Println("Error sending message:", err)
				return
			}
		}
	}
}

func receiveMessages(ctx context.Context, conn *websocket.Conn) {
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
