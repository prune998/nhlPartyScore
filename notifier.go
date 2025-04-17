package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"strings"
	"time"
)

type Notifier interface {
	Notify(game Game, goal Goal, teamID int, followed bool) error
}

// Console Notifier
type ConsoleNotifier struct{}

func (n *ConsoleNotifier) Notify(game Game, goal Goal, teamID int, followed bool) error {
	teamName := ""
	if game.HomeTeam.ID == teamID {
		teamName = game.HomeTeam.Name.DefaultName
	} else {
		teamName = game.AwayTeam.Name.DefaultName
	}

	fmt.Printf("\n🚨 GOAL! 🚨\n")
	fmt.Printf("%s: %s scores at %s of period %d!\n",
		teamName,
		goal.FirstName.DefaultName+" "+goal.LastName.DefaultName,
		goal.TimeInPeriod,
		goal.Period)

	if len(goal.Assists) > 0 {
		var assistNames []string
		for _, assist := range goal.Assists {
			assistNames = append(assistNames, assist.Name.DefaultName)
		}
		fmt.Printf("Assisted by: %s\n", strings.Join(assistNames, ", "))
	}

	fmt.Printf("Score: %s %d - %s %d\n",
		game.HomeTeam.Name, game.HomeTeam.Score,
		game.AwayTeam.Name, game.AwayTeam.Score)

	return nil
}

// Email Notifier
type EmailNotifier struct {
	Config EmailConfig
}

func (n *EmailNotifier) Notify(game Game, goal Goal, teamID int, followed bool) error {
	teamName := ""
	if game.HomeTeam.ID == teamID {
		teamName = game.HomeTeam.Name.DefaultName
	} else {
		teamName = game.AwayTeam.Name.DefaultName
	}

	subject := fmt.Sprintf("Goal Alert: %s scores!", teamName)

	var assistText string
	if len(goal.Assists) > 0 {
		var assistNames []string
		for _, assist := range goal.Assists {
			assistNames = append(assistNames, assist.Name.DefaultName)
		}
		assistText = fmt.Sprintf("Assisted by: %s\n", strings.Join(assistNames, ", "))
	}

	body := fmt.Sprintf("GOAL!\n\n%s: %s scores at %s of period %d!\n%sScore: %s %d - %s %d",
		teamName,
		goal.FirstName.DefaultName+" "+goal.LastName.DefaultName,
		goal.TimeInPeriod,
		goal.Period,
		assistText,
		game.HomeTeam.Name, game.HomeTeam.Score,
		game.AwayTeam.Name, game.AwayTeam.Score)

	message := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s",
		strings.Join(n.Config.To, ","),
		subject,
		body)

	auth := smtp.PlainAuth("", n.Config.Username, n.Config.Password, n.Config.SMTPServer)
	err := smtp.SendMail(
		fmt.Sprintf("%s:%d", n.Config.SMTPServer, n.Config.SMTPPort),
		auth,
		n.Config.From,
		n.Config.To,
		[]byte(message),
	)

	if err != nil {
		return err
	}

	log.Println("Email notification sent")
	return nil
}

// Webhook Notifier
type WebhookNotifier struct {
	Config WebhookConfig
}

func (n *WebhookNotifier) Notify(game Game, goal Goal, teamID int, followed bool) error {
	teamName := ""
	if game.HomeTeam.ID == teamID {
		teamName = game.HomeTeam.Name.DefaultName
	} else {
		teamName = game.AwayTeam.Name.DefaultName
	}

	var assistNames []string
	for _, assist := range goal.Assists {
		assistNames = append(assistNames, assist.Name.DefaultName)
	}

	payload := map[string]interface{}{
		"event":      "goal",
		"team":       teamName,
		"scorer":     goal.FirstName.DefaultName + " " + goal.LastName.DefaultName,
		"period":     goal.Period,
		"time":       goal.TimeInPeriod,
		"assists":    assistNames,
		"home_team":  game.HomeTeam.Name,
		"home_score": game.HomeTeam.Score,
		"away_team":  game.AwayTeam.Name,
		"away_score": game.AwayTeam.Score,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	client := &http.Client{}
	req, err := http.NewRequest(n.Config.Method, n.Config.URL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range n.Config.Headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook request failed with status code: %d", resp.StatusCode)
	}

	log.Println("Webhook notification sent")
	return nil
}

// Hue Light Notifier
type HueNotifier struct {
	Config HueConfig
}

type HueData struct {
	State HueState `json:"state"`
}

type HueState struct {
	On     bool   `json:"on"`
	Bri    int    `json:"bri"`
	Hue    int    `json:"hue"`
	Sat    int    `json:"sat"`
	Effect string `json:"effect"`
}

func (n *HueNotifier) Notify(game Game, goal Goal, teamID int, followed bool) error {
	client := &http.Client{}

	// Send the API request to Philips Hue bridge
	url := fmt.Sprintf("http://%s/api/%s/lights/%d/",
		n.Config.BridgeIP,
		n.Config.Username,
		n.Config.LightID)

	// get current Config
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// Read and parse the current light state
	var currentState *HueData

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	err = json.Unmarshal(body, &currentState)
	if err != nil {
		return err
	}

	// Prepare the light data - we'll flash red for a goal
	alertType := n.Config.FlashType
	if alertType == "" {
		alertType = "lselect" // long flash
	}

	payload := map[string]interface{}{
		"alert": alertType,
		"hue":   n.Config.Hue, // green, 0 Red color (0 is red in Hue's color system)
		"sat":   n.Config.Sat, // Full saturation
		"bri":   n.Config.Bri, // Full brightness
	}
	if !followed {
		payload["hue"] = n.Config.OpponentHue
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url = fmt.Sprintf("http://%s/api/%s/lights/%d/state",
		n.Config.BridgeIP,
		n.Config.Username,
		n.Config.LightID)

	req, err = http.NewRequest("PUT", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Println("Hue light notification sent")
	time.Sleep(15 * time.Second)

	// revert back to previous color
	payload = map[string]interface{}{
		"alert": "none",
		"hue":   currentState.State.Hue,
		"sat":   currentState.State.Sat,
		"bri":   currentState.State.Bri,
		"on":    currentState.State.On,
	}

	jsonPayload, err = json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err = http.NewRequest("PUT", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Println("Hue light notification reverted")
	return nil
}

// Factory function to create appropriate notifier
func NewNotifier(config *Config) []Notifier {
	var notifiers []Notifier

	for _, notifier := range strings.Split(config.NotificationType, ",") {
		fmt.Printf("Creating notifier for %s\n", notifier)
		switch notifier {
		case "email":
			notifiers = append(notifiers, &EmailNotifier{Config: config.Email})
		case "webhook":
			notifiers = append(notifiers, &WebhookNotifier{Config: config.Webhook})
		case "hue":
			notifiers = append(notifiers, &HueNotifier{Config: config.HueConfig})
		default:
			notifiers = append(notifiers, &ConsoleNotifier{})
		}
	}
	return notifiers
}
