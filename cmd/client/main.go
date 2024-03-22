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

	// ルーム名を入力
	fmt.Print("Enter room name: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	room := scanner.Text()

	// サーバーのURLを指定し、ルーム名をクエリパラメータに追加
	serverURL := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws", RawQuery: "room=" + room}

	// WebSocketに接続
	conn, _, err := websocket.DefaultDialer.Dial(serverURL.String(), nil)
	if err != nil {
		fmt.Println("Error connecting to WebSocket server:", err)
		return
	}
	defer conn.Close()

	// メッセージ受信用のゴルーチンを起動
	go receiveMessages(conn)

	// ユーザーからのメッセージを読み込んで送信
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
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error receiving message:", err)
			break
		}
		fmt.Printf("Received message: %s\n", string(msg))
	}
}
