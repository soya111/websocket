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

	// ctrl+c cancels the context and tells scanner to EOF
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		fmt.Printf("ctx.Done() received: %v\n", ctx.Err())
		err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			fmt.Println("Error sending close message:", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := receiveMessages(ctx, conn)
		fmt.Println("Error from receiveMessages:", err)
		// context not done means connection closed by server, so output press enter to exit sendMessages loop
		if ctx.Err() == nil {
			fmt.Println("Press enter to exit")
		}
		stop()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := sendMessages(ctx, conn)
		fmt.Println("Error from sendMessages:", err)
	}()

	wg.Wait()
	fmt.Println("Client shutting down successfully")
}

func sendMessages(ctx context.Context, conn *websocket.Conn) error {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context Done: %w", ctx.Err())
		default:
			if !scanner.Scan() {
				return fmt.Errorf("scanner Scan returned false: %w", scanner.Err())
			}
			msg := scanner.Text()
			err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				return fmt.Errorf("Write message: %w", err)
			}
		}
	}
}

func receiveMessages(ctx context.Context, conn *websocket.Conn) error {
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context Done: %w", ctx.Err())
		default:
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				return fmt.Errorf("Read message: message type: %d, error: %w", msgType, err)
			}
			fmt.Printf("Read message: %s\n", string(msg))
		}
	}
}
