package main

import (
	"bufio"
	"encoding/json"
	"galcone/src/galcone/container"
	"galcone/src/galcone/messages/incoming"
	"galcone/src/galcone/models"
	"log"
	"net"
	"os"
)

const (
	//ConnHost = "localhost"
	ConnPort = "8081"
	ConnType = "tcp"
)

type handler func(*models.Player, *container.GamesContainer, *json.RawMessage)

var RequestHandlers = map[string]handler{
	"player_ready": incoming.HandlePlayerReadyRequest,
	"join":         incoming.HandlePlayerJoinRequest,
	"leave":        incoming.HandlePlayerLeaveRequest,
	"send_ships":   incoming.HandleSendShipsRequest,
}

func main() {
	listener, err := net.Listen(ConnType, ":"+ConnPort)
	if err != nil {
		log.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer listener.Close()
	log.Println("Listening on port:", ConnPort)

	gameContainer := container.NewGamesContainer()
	go gameContainer.Run()

	for {
		conn, err := listener.Accept()
		log.Print(conn)
		if err != nil {
			log.Println("Error accepting connection:", err.Error())
			os.Exit(1)
		}

		log.Printf("New connection established with player: %v", conn.RemoteAddr())

		player := models.Player{
			Connection: &conn,
		}
		log.Print(player)
		go handleRequest(gameContainer, &player)
	}
}

func handleRequest(container *container.GamesContainer, player *models.Player) {
	log.Printf("Handling requests for player ID: %v", player.Id)
	for {
		msgType, payload, err := readMessage(player.Connection)
		if err {
			log.Printf("Connection closed for player with ID: %d and session ID: %d", player.Id, player.SessionId)
			return
		}

		log.Printf("Received message type: %s from player ID: %d", msgType, player.Id)

		requestHandler := RequestHandlers[msgType]
		if requestHandler != nil {
			log.Printf("Handling request type: %s for player ID: %d", msgType, player.Id)
			requestHandler(player, container, payload)
		} else {
			log.Printf("No handler found for request type: %s", msgType)
		}
	}
}

func readMessage(conn *net.Conn) (string, *json.RawMessage, bool) {
	log.Print(conn)
    request, err := bufio.NewReader(*conn).ReadString('\n')
    if err != nil {
        if err.Error() == "EOF" {
            log.Println("Connection closed by client (EOF)")
        } else {
            log.Printf("Error reading: %s", err.Error())
        }
        return "", nil, true
    }

    log.Printf("Received message: %s", request)

    var payload json.RawMessage
    msg := models.Message{
        Payload: &payload,
    }
    if err := json.Unmarshal([]byte(request), &msg); err != nil {
        log.Printf("Error unmarshalling message: %s", err.Error())
        return "", nil, true
    }
    return msg.Type, &payload, false
}
