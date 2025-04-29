package outgoing

import (
	"encoding/json"
	"galcone/src/galcone/models"
	"log"
	"net"
)

const (
	JoinAcceptedMessageType    = "join_accepted"
	PlayerJoinedMessageType    = "player_joined"
	PlayerReadyMessageType     = "player_ready"
	PlayerLeftMessageType      = "player_left"
	PlayerKickedMessageType    = "player_kicked"
	ShipsSentResponseMessageType = "ships_sent"
)

type PlanetInResponse struct {
	Id         int  `json:"id"`
	Size       int  `json:"size"`
	Population int  `json:"population"`
	PosX       int  `json:"position_x"`
	PosY       int  `json:"position_y"`
	PlayerId   *int `json:"player_id"`
}

type PlayerReadyResponse struct {
	Login string `json:"login"`
}

type PlayerKickedResponse struct {
	Reason string `json:"reason"`
}

type JoinAcceptedResponse struct {
	PlayerId         int                 `json:"player_id"`
	SessionId        int                 `json:"session_id"`
	StartingPlanetId int                 `json:"starting_planet_id"`
	Planets          []*PlanetInResponse `json:"planets"`
	GrowthRate       float64             `json:"population_growth_rate"`
}

type PlayerJoinedResponse struct {
	PlayerName       string `json:"name"`
	StartingPlanetId int    `json:"starting_planet_id"`
}

type PlayerLeftResponse struct {
	PlayerName string `json:"player_name"`
}

type ShipsSentResponse struct {
	FromPlanetId      int   `json:"from"`
	ToPlanetId        int   `json:"to"`
	Amount            int   `json:"amount"`
	GroupId           int   `json:"group_id"`
	ArrivalTimestamp  int64 `json:"arrival_timestamp"`
}

func SendJsonResponse(message *models.Message, connection *net.Conn) {
	jsonBody, err := json.Marshal(message)
	if err != nil {
		log.Printf("[outgoing] Failed to marshal message of type '%s': %v", message.Type, err)
		return
	}

	_, err = (*connection).Write(jsonBody)
	if err != nil {
		log.Printf("[outgoing] Failed to send message of type '%s': %v", message.Type, err)
		return
	}

	log.Printf("[outgoing] Sent message of type '%s'", message.Type)
}

func NotifyPlayerLeft(session *models.GameSession, leftPlayer *models.Player) {
	log.Printf("[outgoing] Notifying players that '%s' left session %d", leftPlayer.Login, session.Id)

	msg := &models.Message{
		Type: PlayerLeftMessageType,
		Payload: &PlayerLeftResponse{
			PlayerName: leftPlayer.Login,
		},
	}
	notifyAllExceptSender(msg, session, leftPlayer)
}

func NotifyPlayerJoined(session *models.GameSession, joinedPlayer *models.Player, startingPlanet *models.Planet) {
	log.Printf("[outgoing] Notifying that player '%s' joined session %d", joinedPlayer.Login, session.Id)

	notifyJoinedPlayer(session, joinedPlayer, startingPlanet)
	notifyOtherPlayers(session, joinedPlayer, startingPlanet)
}

func notifyJoinedPlayer(session *models.GameSession, joinedPlayer *models.Player, startingPlanet *models.Planet) {
	log.Printf("[outgoing] Sending join accepted to '%s'", joinedPlayer.Login)

	planetsInResponse := make([]*PlanetInResponse, len(session.Planets))
	for key, planet := range session.Planets {
		planetsInResponse[key] = convertPlanetToResponseFormat(*planet)
	}

	joinAcceptedMsg := models.Message{
		Type: JoinAcceptedMessageType,
		Payload: &JoinAcceptedResponse{
			PlayerId:         joinedPlayer.Id,
			SessionId:        session.Id,
			Planets:          planetsInResponse,
			StartingPlanetId: startingPlanet.Id,
			GrowthRate:       6.316,
		},
	}

	SendJsonResponse(&joinAcceptedMsg, joinedPlayer.Connection)
}

func notifyOtherPlayers(session *models.GameSession, joinedPlayer *models.Player, startingPlanet *models.Planet) {
	log.Printf("[outgoing] Notifying other players that '%s' joined", joinedPlayer.Login)

	playerJoinedMsg := &models.Message{
		Type: PlayerJoinedMessageType,
		Payload: &PlayerJoinedResponse{
			PlayerName:       joinedPlayer.Login,
			StartingPlanetId: startingPlanet.Id,
		},
	}

	notifyAllExceptSender(playerJoinedMsg, session, joinedPlayer)
}

func notifyAllExceptSender(msg *models.Message, session *models.GameSession, sender *models.Player) {
	for _, player := range session.Players {
		if player.Id != sender.Id {
			log.Printf("[outgoing] Sending message of type '%s' to player '%s'", msg.Type, player.Login)
			SendJsonResponse(msg, player.Connection)
		}
	}
}

func convertPlanetToResponseFormat(planet models.Planet) *PlanetInResponse {
	planetInResponse := &PlanetInResponse{
		Id:         planet.Id,
		Size:       planet.Size,
		Population: planet.Population,
		PosX:       planet.Coordx,
		PosY:       planet.Coordy,
	}

	if planet.Player != nil {
		planetInResponse.PlayerId = &planet.Player.Id
	}

	return planetInResponse
}
