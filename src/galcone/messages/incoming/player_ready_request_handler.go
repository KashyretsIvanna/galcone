package incoming

import (
	"encoding/json"
	"galcone/src/galcone/container"
	"galcone/src/galcone/messages/outgoing"
	"galcone/src/galcone/models"
	"log"
)

type PlayerReadyRequest struct {
	SessionId int
	PlayerId  int
}

func HandlePlayerReadyRequest(player *models.Player, container *container.GamesContainer, payload *json.RawMessage) {
	// Log the incoming request
	log.Printf("Received PlayerReadyRequest: SessionId=%d, PlayerId=%d", player.SessionId, player.Id)

	// Attempt to unmarshal the payload into the requestBody
	var requestBody PlayerReadyRequest
	err := json.Unmarshal(*payload, &requestBody)
	if err != nil {
		log.Printf("Error unmarshalling payload: %v", err)
		panic("Wrong json!")
	}

	log.Printf("Successfully unmarshalled PlayerReadyRequest: %+v", requestBody)

	// Set the player as ready in the container
	updatedPlayer := container.SetPlayerReady(player.SessionId, player.Id)
	log.Printf("Player %d (%s) set to ready in session %d", player.Id, updatedPlayer.Login, player.SessionId)

	// Update the session status
	container.UpdateSessionStatus(requestBody.SessionId)
	log.Printf("Session %d status updated", requestBody.SessionId)

	// Get all players from the session and send readiness responses
	players := container.GetPlayersFromSession(requestBody.SessionId)
	for _, player := range players {
		if player.Ready {
			// Send the readiness response to players that are ready
			response := &outgoing.PlayerReadyResponse{Login: updatedPlayer.Login}
			msg := &models.Message{
				Type:    outgoing.PlayerReadyMessageType,
				Payload: response,
			}
			log.Printf("Sending PlayerReadyResponse to player %d (%s)", player.Id, player.Login)
			outgoing.SendJsonResponse(msg, player.Connection)
		}
	}
}
