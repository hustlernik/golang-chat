package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type ChatHub struct {
  clients    map[*websocket.Conn]bool
  broadcast  chan string
	mu         sync.Mutex
}

func newChatHub() *ChatHub {
	return &ChatHub{
    clients: make(map[*websocket.Conn]bool),
    broadcast: make(chan string),
  }
}

func (hub *ChatHub) run() {
	for {
    message := <-hub.broadcast

    hub.mu.Lock()
    for client := range hub.clients {
      err := client.WriteMessage(websocket.TextMessage, []byte(message))
      if err!= nil {
        fmt.Println("Error in writing to client:", err)
        delete(hub.clients, client)
        client.Close()
      }
    }
    hub.mu.Unlock()
  }
}

func (hub *ChatHub) handleConnection(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading connection:", err)
		return
	}
	defer conn.Close()


	hub.mu.Lock()
	hub.clients[conn] = true
	hub.mu.Unlock()


	for {
		_, message, err := conn.ReadMessage()
		if err != nil {

			hub.mu.Lock()
			delete(hub.clients, conn)
			hub.mu.Unlock()
			break
		}


		hub.broadcast <- string(message)
	}
}



func main() {
	 hub:= newChatHub()

	 go hub.run()

	 http.HandleFunc("/ws",hub.handleConnection)

	 fmt.Println("chat backend running on :8080")

	 if err:= http.ListenAndServe(":8080",nil); err!=nil {
     fmt.Println("Server error: ",err)
	 }
}