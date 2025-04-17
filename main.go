package main

import (
	"flag"
	"log"
	"time"
)

func main() {
	configFile := flag.String("config", "config.yaml", "Path to the configuration file")
	flag.Parse()

	config, err := LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	notifiers := NewNotifier(config)

	log.Printf("Starting NHL Goal Watcher for team ID: %d", config.TeamID)
	// log.Printf("Polling interval: %d seconds", config.PollInterval)
	log.Printf("Notification types: %s", config.NotificationType)

	// Keep track of goals we've already seen
	var lastGoalCount int
	var currentGame *Game
	var initialized bool
	for {
		// Fetch current scores
		scores, err := fetchNHLScores()
		if err != nil {
			log.Printf("Error fetching NHL scores: %v", err)
			time.Sleep(time.Second * config.PollInterval)
			continue
		}

		// Find the game for our team
		game := findTeamGame(scores, config.TeamID)
		if game == nil {
			if currentGame != nil {
				log.Println("No active game found for the specified team.")
				currentGame = nil
				lastGoalCount = 0
			}
			time.Sleep(time.Second * config.PollInterval)
			continue
		}

		// If this is a new game or first time seeing this game
		if currentGame == nil || currentGame.ID != game.ID {
			log.Printf("Found game: %s(%d) vs %s(%d)", game.HomeTeam.Name.DefaultName, game.HomeTeam.Score, game.AwayTeam.Name.DefaultName, game.AwayTeam.Score)
			currentGame = game
			lastGoalCount = 0
		}

		// Only check for goals if the game is in progress
		if game.GameState != "FINAL" && game.GameState != "OFF" && game.GameState != "PREVIEW" {
			currentGoalCount := len(game.Goals)

			// Check if we have new goals
			if currentGoalCount > lastGoalCount {
				// For each new goal, check if it's for our team
				for i := lastGoalCount; i < currentGoalCount; i++ {
					goal := game.Goals[i]

					// Check if the goal is for our team
					var goalTeamID int
					if goal.TeamAbbrev == game.HomeTeam.Abbrev {
						goalTeamID = game.HomeTeam.ID
					} else {
						goalTeamID = game.AwayTeam.ID
					}

					followed := goalTeamID == config.TeamID
					if initialized {
						for _, notifier := range notifiers {
							err = notifier.Notify(*game, goal, config.TeamID, followed)
							if err != nil {
								log.Printf("Error sending notification: %v", err)
							}
						}
					}
				}
				initialized = true
				lastGoalCount = currentGoalCount
			}

		} else if game.GameState == "FINAL" && currentGame != nil && currentGame.GameState != "FINAL" {
			// Game just ended
			log.Printf("Game ended: %s %d - %s %d",
				game.HomeTeam.Name.DefaultName, game.HomeTeam.Score,
				game.AwayTeam.Name.DefaultName, game.AwayTeam.Score)
			currentGame = game
		}

		time.Sleep(config.PollInterval)
	}
}
