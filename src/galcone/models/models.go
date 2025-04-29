package models

import (
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type    string
	Payload interface{}
}

type Planet struct {
	Id         int
	Size       int
	Population int
	Coordx     int
	Coordy     int
	Player     *Player
}

type Group struct {
	Id           int
	Amount       int
	Coordx       int //only if we implement redirect
	Coordy       int //only if we implement redirect
	TargetPlanet Planet
	SourcePlanet Planet
	SourceGroup  *Group //only if we implement redirect
	ArrivalTime  time.Time
	Player       Player
}

type Player struct {
	Id         int
	SessionId  int
	Connection *websocket.Conn
	Login      string
	Ready      bool
}

type GameSession struct {
	Id              int
	Active          bool
	MaxPlayersCount int
	Planets         []*Planet
	Groups          []*Group
	Players         map[int]*Player
}

func (session *GameSession) IsFull() bool {
	return len(session.Players) == session.MaxPlayersCount
}

func (session *GameSession) AddPlayerToSession(player *Player) bool {
	if session.IsFull() {
		return false
	}

	player.Id = len(session.Players)
	player.SessionId = session.Id
	session.Players[player.Id] = player
	return true
}

func (session *GameSession) GetFreePlanet() *Planet {
	for _, planet := range session.Planets {
		if planet.Player == nil {
			return planet
		}
	}

	return nil
}

func (session *GameSession) GetPlanetById(planetId int) *Planet {
	for _, planet := range session.Planets {
		if planet.Id == planetId {
			return planet
		}
	}

	return nil
}

func (s *GameSession) CheckWinner() *Player {
	if len(s.Planets) == 0 {
		return nil
	}

	var candidateOwner *Player
	for _, planet := range s.Planets {
		if planet.Player == nil {
			return nil // Empty planet — game not over
		}
		if candidateOwner == nil {
			candidateOwner = planet.Player
		} else if candidateOwner.Id != planet.Player.Id {
			return nil // Different owners — game not over
		}
	}
	return candidateOwner // All planets belong to the same player
}

func (s *GameSession) StartPopulationGrowth() {
	ticker := time.NewTicker(3 * time.Second) // every 3 seconds
	go func() {
		for {
			select {
			case <-ticker.C:
				for _, planet := range s.Planets {
					if planet.Player != nil {
						// Growth amount based on size
						growthRate := 1 + planet.Size/10 // you can adjust this formula
						planet.Population += growthRate
					}
				}
			}
		}
	}()
}

func (session *GameSession) RemovePlayerFromSession(player *Player) {
	if session.Active {
		player.Ready = false
	} else {
		delete(session.Players, player.Id)
	}

	(*player.Connection).Close()
}

func (p *Planet) ReceiveShips(fromPlayer *Player, shipsAmount int) {
	log.Printf("Planet %d receiving %d ships from Player %d. Current Population: %d, Current Owner: %v",
		p.Id, shipsAmount, fromPlayer.Id, p.Population, playerInfo(p.Player))

	if p.Player != nil && p.Player.Id == fromPlayer.Id {
		// Reinforcing own planet
		p.Population += shipsAmount
		log.Printf("Reinforcement: Planet %d new Population: %d", p.Id, p.Population)
	} else {
		if shipsAmount >= p.Population {
			// Planet captured because the attacking amount is greater than or equal to population
			oldOwnerId := -1
			if p.Player != nil {
				oldOwnerId = p.Player.Id
			}
			p.Player = fromPlayer
			p.Population = shipsAmount - p.Population
			log.Printf("Planet %d captured by Player %d (was owned by Player %d). New Population: %d",
				p.Id, fromPlayer.Id, oldOwnerId, p.Population)
		} else {
			// Defense successful, but planet population decreases
			p.Population -= shipsAmount
			log.Printf("Defense successful: Planet %d remaining Population: %d", p.Id, p.Population)
		}
	}
}

func playerInfo(p *Player) string {
	if p == nil {
		return "none"
	}
	return fmt.Sprintf("%d", p.Id)
}
