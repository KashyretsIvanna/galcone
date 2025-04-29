package container

import (
	"galcone/src/galcone/messages/outgoing"
	"galcone/src/galcone/models"
	"log"
)

type GamesContainer struct {
	GameSessions map[int] *models.GameSession
	JoinQueue    chan *models.Player
	LeaveQueue   chan *models.Player
}

func NewGamesContainer() *GamesContainer {
	log.Println("Initializing GamesContainer...")
	return &GamesContainer {
		JoinQueue: make(chan *models.Player),
		LeaveQueue: make(chan *models.Player),
		GameSessions: make(map[int] *models.GameSession),
	}
}

func (container *GamesContainer) Run() {
	log.Println("Running the GamesContainer...")
	for {
		select {
		case player := <-container.JoinQueue:
			log.Printf("Processing join request for player %v...", player.Id)
			session := container.findAvailableToJoinSession()
			session.AddPlayerToSession(player)
			freePlanet := session.GetFreePlanet()
			freePlanet.Player = player
			log.Printf("Player %v joined session %v on planet %v", player.Id, session.Id, freePlanet.Id)
			outgoing.NotifyPlayerJoined(session, player, freePlanet)
		case player := <-container.LeaveQueue:
			log.Printf("Processing leave request for player %v...", player.Id)
			session := container.GetGameSessionById(player.SessionId)
			session.RemovePlayerFromSession(player)
			log.Printf("Player %v left session %v", player.Id, session.Id)
			outgoing.NotifyPlayerLeft(session, player)
		}
	}
}

func (container *GamesContainer) findAvailableToJoinSession() *models.GameSession {
	log.Println("Searching for an available session to join...")
	for _, session := range container.GameSessions {
		if !session.IsFull() {
			log.Printf("Found available session %v for player.", session.Id)
			return session
		}
	}

	// No available session, creating a new one
	log.Println("No available sessions found, creating a new session...")
	newSession := &models.GameSession{
		Id: len(container.GameSessions),
		MaxPlayersCount: 2,
		Players: make(map[int] *models.Player),
		Planets: container.generatePlanets(),
	}
	container.GameSessions[newSession.Id] = newSession
	log.Printf("New session %v created.", newSession.Id)
	return newSession
}

func (container *GamesContainer) generatePlanets() []*models.Planet {
	log.Println("Generating new planets for the session...")
	planets := []*models.Planet{
		&models.Planet{Id: 1, Size: 6, Coordx: 1, Coordy: 1, Population: 45, Player: nil},
		&models.Planet{Id: 2, Size: 6, Coordx: 9, Coordy: 9, Population: 45, Player: nil},
		&models.Planet{Id: 3, Size: 4, Coordx: 2, Coordy: 8, Population: 30, Player: nil},
		&models.Planet{Id: 4, Size: 4, Coordx: 8, Coordy: 2, Population: 40, Player: nil},
		&models.Planet{Id: 5, Size: 4, Coordx: 4, Coordy: 4, Population: 50, Player: nil},
	}
	log.Printf("Generated %v planets.", len(planets))
	return planets
}
