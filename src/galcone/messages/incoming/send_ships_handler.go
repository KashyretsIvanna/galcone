package incoming

import (
	"encoding/json"
	"galcone/src/galcone/container"
	"galcone/src/galcone/messages/outgoing"
	"galcone/src/galcone/models"
	"log"
	"math"
	"time"
)

type SendShipsRequest struct {
	FromPlanetId int `json:"from"`
	ToPlanetId   int `json:"to"`
}

func HandleSendShipsRequest(player *models.Player, container *container.GamesContainer, payload *json.RawMessage) {
	// Log incoming request
	log.Printf("Received SendShipsRequest: PlayerId=%d SessionId=%d Payload=%s", player.Id, player.SessionId, string(*payload))

	// Unmarshal the payload into the request body
	var requestBody SendShipsRequest
	err := json.Unmarshal(*payload, &requestBody)
	if err != nil {
		log.Printf("Error unmarshalling payload: %v", err)
		panic("Wrong json!")
	}

	// Get the game session by player session ID
	gameSession := container.GetGameSessionById(player.SessionId)
	if gameSession == nil {
		log.Printf("Game session not found for PlayerId=%d", player.Id)
		return
	}

	// Check if the game session is active
	if !gameSession.Active {
		log.Printf("Session is not active for PlayerId=%d", player.Id)
		return
	}

	// Get the source planet by ID
	sourcePlanet := gameSession.GetPlanetById(requestBody.FromPlanetId)
	if sourcePlanet == nil {
		log.Printf("Source planet not found: PlanetId=%d for PlayerId=%d", requestBody.FromPlanetId, player.Id)
		return
	}

	// Check if the player owns the source planet
	if sourcePlanet.Player == nil || sourcePlanet.Player.Id != player.Id {
		log.Printf("Player %d is not the owner of source planet %d", player.Id, requestBody.FromPlanetId)
		return
	}

	// Get the target planet by ID
	targetPlanet := gameSession.GetPlanetById(requestBody.ToPlanetId)
	if targetPlanet == nil {
		log.Printf("Target planet not found: PlanetId=%d for PlayerId=%d", requestBody.ToPlanetId, player.Id)
		return
	}

	// Log planet details before sending ships
	log.Printf("Before sending ships: Source Planet %d Population: %d, Target Planet %d Population: %d",
		sourcePlanet.Id, sourcePlanet.Population, targetPlanet.Id, targetPlanet.Population)

	// Calculate the distance between the source and target planets
	xDistance := math.Pow(float64(sourcePlanet.Coordx-targetPlanet.Coordx), 2)
	yDistance := math.Pow(float64(sourcePlanet.Coordy-targetPlanet.Coordy), 2)
	distanceBetweenPlanets := math.Sqrt(xDistance + yDistance)

	amountToSend := sourcePlanet.Population / 2
	if amountToSend <= 0 {
		log.Printf("Cannot send ships from empty planet")
		return
	}
	sourcePlanet.Population -= amountToSend
	// Create a group for the ships and set the arrival time
	group := &models.Group{
		Id:           len(gameSession.Groups),
		Amount:       amountToSend,
		SourcePlanet: *sourcePlanet,
		TargetPlanet: *targetPlanet,
		ArrivalTime:  time.Now().Add(time.Second * time.Duration(distanceBetweenPlanets)),
		Player:       *player,
	}
	gameSession.Groups = append(gameSession.Groups, group)

	// Log the ship sending
	log.Printf("Sending ships: GroupId=%d FromPlanetId=%d ToPlanetId=%d Amount=%d ArrivalTime=%s",
		group.Id, group.SourcePlanet.Id, group.TargetPlanet.Id, group.Amount, group.ArrivalTime)


	// Log planet details after sending ships
	log.Printf("After sending ships: Source Planet %d Population: %d, Target Planet %d Population: %d",
		sourcePlanet.Id, sourcePlanet.Population, targetPlanet.Id, targetPlanet.Population)

	// Send responses to all players in the game session
	for _, p := range gameSession.Players {
		response := &outgoing.ShipsSentResponse{
			GroupId:          group.Id,
			FromPlanetId:     group.SourcePlanet.Id,
			ToPlanetId:       group.TargetPlanet.Id,
			Amount:           group.Amount,
			ArrivalTimestamp: group.ArrivalTime.Unix(),
		}

		msg := &models.Message{
			Type:    outgoing.ShipsSentResponseMessageType,
			Payload: response,
		}

		log.Printf("Sending ShipsSentResponse to PlayerId=%d", p.Id)
		outgoing.SendJsonResponse(msg, p.Connection)
	}

	// Schedule ships arrival
	time.AfterFunc(time.Until(group.ArrivalTime), func() {
		log.Printf("Ships arrived: GroupId=%d FromPlanetId=%d ToPlanetId=%d", group.Id, group.SourcePlanet.Id, group.TargetPlanet.Id)

		// Update target planet's population
		targetPlanet := gameSession.GetPlanetById(group.TargetPlanet.Id)
		if targetPlanet != nil {
			targetPlanet.ReceiveShips(&group.Player, group.Amount)

			log.Printf("After arrival: Target Planet %d Population: %d", targetPlanet.Id, targetPlanet.Population)
		}

		// Notify players that ships have landed
		for _, p := range gameSession.Players {
			response := map[string]interface{}{
				"groupId":      group.Id,
				"fromPlanetId": group.SourcePlanet.Id,
				"toPlanetId":   group.TargetPlanet.Id,
				"amount":       group.Amount,
			}

			msg := &models.Message{
				Type:    "ships_arrived",
				Payload: response,
			}

			outgoing.SendJsonResponse(msg, p.Connection)
			log.Printf("[outgoing] Sent message of type 'ships_arrived' to PlayerId=%d", p.Id)
		}

		// (Optional) Check if someone won after ships arrived
		winner := gameSession.CheckWinner()
		if winner != nil {
			// Send victory message to all players
			for _, p := range gameSession.Players {
				msg := &models.Message{
					Type: "game_over",
					Payload: map[string]interface{}{
						"winnerId": winner.Id,
					},
				}
				outgoing.SendJsonResponse(msg, p.Connection)
			}
			log.Printf("Player %d wins the game!", winner.Id)

			// Optional: Close game session, cleanup, etc.
		}

		// (Optional) Remove the group from session.Groups if you want to clean memory
	})
}
