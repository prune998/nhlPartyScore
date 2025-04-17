package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type GameScores struct {
	Games []Game `json:"games"`
}

type Game struct {
	ID               int              `json:"id"`
	GameDate         string           `json:"gameDate"`
	GameState        string           `json:"gameState"`
	HomeTeam         Team             `json:"homeTeam"`
	AwayTeam         Team             `json:"awayTeam"`
	PeriodDescriptor PeriodDescriptor `json:"periodDescriptor"`
	GameOutcome      *GameOutcome     `json:"gameOutcome,omitempty"`
	Goals            []Goal           `json:"goals"`
}

type Team struct {
	ID     int    `json:"id"`
	Name   Name   `json:"name"`
	Abbrev string `json:"abbrev"`
	Score  int    `json:"score"`
	SoG    int    `json:"sog"`
	Logo   string `json:"logo"`
}

type Name struct {
	DefaultName string `json:"default"`
	NameFR      string `json:"fr"`
}

type PeriodDescriptor struct {
	Number int    `json:"number"`
	Period string `json:"period"`
}

type GameOutcome struct {
	LastPeriodType string `json:"lastPeriodType"`
}

type Goal struct {
	Period           int              `json:"period"`
	PeriodDescriptor PeriodDescriptor `json:"periodDescriptor"`
	TimeInPeriod     string           `json:"timeInPeriod"`
	TeamAbbrev       string           `json:"teamAbbrev"`
	PlayerID         int              `json:"playerId"`
	Assists          []Scorer         `json:"assists"`
	StrengthCode     string           `json:"strengthCode"`
	EmptyNet         bool             `json:"emptyNet"`
	EventID          string           `json:"eventId"`
	Name             Name             `json:"name"`
	FirstName        Name             `json:"firstName"`
	LastName         Name             `json:"lastName"`
}

type Scorer struct {
	ID   int  `json:"playerId"`
	Name Name `json:"name"`
}

func fetchNHLScores() (*GameScores, error) {
	today := time.Now().Local().Format("2006-01-02")
	url := fmt.Sprintf("https://api-web.nhle.com/v1/score/%s", today)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var scores GameScores
	err = json.Unmarshal(body, &scores)
	if err != nil {
		return nil, err
	}

	return &scores, nil
}

// Find a team's game for today
func findTeamGame(scores *GameScores, teamID int) *Game {
	for _, game := range scores.Games {
		if game.HomeTeam.ID == teamID || game.AwayTeam.ID == teamID {
			return &game
		}
	}
	return nil
}

// Check if this is a new goal we haven't seen before
func isNewGoal(savedGoals, currentGoals []Goal, teamID int) (bool, Goal) {
	if len(currentGoals) > len(savedGoals) {
		// Get the newest goal
		newGoal := currentGoals[len(currentGoals)-1]

		// Make sure it's for our team
		var newTeamID int
		if newGoal.TeamAbbrev == "BOS" { // You'd need to map team abbreviation to ID properly
			newTeamID = 6 // Example for Boston Bruins
		}

		if newTeamID == teamID {
			return true, newGoal
		}
	}
	return false, Goal{}
}
