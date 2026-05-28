package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
	"os"
	"github.com/gorilla/websocket"
)

type Message struct {
	Type    string `json:"type"`
	User    string `json:"user"`
	Content string `json:"content"`
	Time    string `json:"time"`
}

type Client struct {
	Conn *websocket.Conn
	Name string
}

var clients = make(map[*Client]bool)

var broadcast = make(chan Message)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func wsHandler(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
		return
	}

	var username string

	err = conn.ReadJSON(&username)

	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		Conn: conn,
		Name: username,
	}

	clients[client] = true

	joinMessage := Message{
		Type:    "system",
		User:    "System",
		Content: username + " joined the chat",
		Time:    time.Now().Format("03:04 PM"),
	}

	broadcast <- joinMessage

	broadcast <- Message{
		Type:    "online",
		Content: fmt.Sprintf("%d users online", len(clients)),
		Time:    time.Now().Format("03:04 PM"),
	}

	for {

		var msg Message

		err := conn.ReadJSON(&msg)

		if err != nil {

			delete(clients, client)

			leaveMessage := Message{
				Type:    "system",
				User:    "System",
				Content: client.Name + " left the chat",
				Time:    time.Now().Format("03:04 PM"),
			}

			broadcast <- leaveMessage

			broadcast <- Message{
				Type:    "online",
				Content: fmt.Sprintf("%d users online", len(clients)),
				Time:    time.Now().Format("03:04 PM"),
			}

			conn.Close()

			break
		}

		msg.Time = time.Now().Format("03:04 PM")

		broadcast <- msg
	}
}

func handleMessages() {

	for {

		msg := <-broadcast

		for client := range clients {

			err := client.Conn.WriteJSON(msg)

			if err != nil {

				client.Conn.Close()

				delete(clients, client)
			}
		}
	}
}

func main() {

	fs := http.FileServer(http.Dir("./static"))

	http.Handle("/", fs)

	http.HandleFunc("/ws", wsHandler)

	go handleMessages()

	log.Println("Server running on :3000")

	port := os.Getenv("PORT")

	if port == "" {
		port = "3000"
	}
	
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
