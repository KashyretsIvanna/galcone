package main

import (
	"encoding/json"
	"galcone/src/galcone/container"
	"galcone/src/galcone/messages/incoming"
	"galcone/src/galcone/models"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

const (
	WebSocketPort = ":3000"
)

type handler func(*models.Player, *container.GamesContainer, *json.RawMessage)

var RequestHandlers = map[string]handler{
	"player_ready": incoming.HandlePlayerReadyRequest,
	"join":         incoming.HandlePlayerJoinRequest,
	"leave":        incoming.HandlePlayerLeaveRequest,
	"send_ships":   incoming.HandleSendShipsRequest,
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // allow all origins for simplicity
	},
}

func main() {
	gameContainer := container.NewGamesContainer()
	go gameContainer.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Upgrade error:", err)
			return
		}

		player := &models.Player{
			Connection: conn, // You will need to update your Player model to add `WsConn *websocket.Conn`
		}
		log.Printf("New WebSocket connection: %v", conn.RemoteAddr())
		go handleRequest(gameContainer, player)
	})

	log.Println("WebSocket server listening on port", WebSocketPort)
	err := http.ListenAndServe(WebSocketPort, nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func handleRequest(container *container.GamesContainer, player *models.Player) {
	for {
		_, message, err := player.Connection.ReadMessage()
		if err != nil {
			log.Printf("Connection closed for player ID %d: %v", player.Id, err)
			return
		}

		var payload json.RawMessage
		msg := models.Message{
			Payload: &payload,
		}
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Println("Error unmarshalling message:", err)
			continue
		}

		requestHandler := RequestHandlers[msg.Type]
		if requestHandler != nil {
			requestHandler(player, container, &payload)
		} else {
			log.Printf("No handler for message type: %s", msg.Type)
		}
	}
}
