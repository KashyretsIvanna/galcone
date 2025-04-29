package incoming

import (
	"encoding/json"
	"galcone/src/galcone/container"
	"galcone/src/galcone/models"
	"log"
)

type PlayerJoinRequest struct {
	PlayerName string `json:"player_name"`
}

func HandlePlayerJoinRequest(player *models.Player, container *container.GamesContainer, payload *json.RawMessage) {
	// Log the incoming request
	log.Printf("Received PlayerJoinRequest for player_name: %s", player.Login)

	// Attempt to unmarshal the payload into the request
	var request PlayerJoinRequest
	if err := json.Unmarshal(*payload, &request); err != nil {
		log.Printf("Error unmarshalling payload: %v", err)
		panic("Unable to parse json!")
	}

	// Log successful unmarshalling
	log.Printf("Successfully unmarshalled PlayerJoinRequest: PlayerName=%s", request.PlayerName)

	// Assign the player name from the request and add player to the join queue
	player.Login = request.PlayerName
	log.Printf("Player %s (ID: %d) is joining the queue", player.Login, player.Id)
	container.JoinQueue <- player
}

func HandlePlayerLeaveRequest(player *models.Player, container *container.GamesContainer, payload *json.RawMessage) {
	// Log the player leaving request
	if player.Login == "" {
		log.Printf("Player is not logged in, skipping leave request")
		return
	}

	log.Printf("Player %s (ID: %d) is leaving the queue", player.Login, player.Id)
	container.LeaveQueue <- player
}
