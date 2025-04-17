package main

import (
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	TeamID           int           `yaml:"team_id"`
	PollInterval     time.Duration `yaml:"poll_interval_seconds"`
	NotificationType string        `yaml:"notification_type"` // console, email, webhook, hue
	Email            EmailConfig   `yaml:"email,omitempty"`
	Webhook          WebhookConfig `yaml:"webhook,omitempty"`
	HueConfig        HueConfig     `yaml:"hue,omitempty"`
}

type EmailConfig struct {
	SMTPServer string   `yaml:"smtp_server"`
	SMTPPort   int      `yaml:"smtp_port"`
	Username   string   `yaml:"username"`
	Password   string   `yaml:"password"`
	From       string   `yaml:"from"`
	To         []string `yaml:"to"`
}

type WebhookConfig struct {
	URL     string            `yaml:"url"`
	Method  string            `yaml:"method"`
	Headers map[string]string `yaml:"headers"`
}

type HueConfig struct {
	BridgeIP    string `yaml:"bridge_ip"`
	Username    string `yaml:"username"`
	LightID     int    `yaml:"light_id"`
	FlashType   string `yaml:"flash_type"`   // "short" or "long"
	Hue         int    `yaml:"hue"`          // green, 0 Red color (0 is red in Hue's color system)
	OpponentHue int    `yaml:"opponent_hue"` // green, 0 Red color (0 is red in Hue's color system)
	Sat         int    `yaml:"sat"`          // Full saturation
	Bri         int    `yaml:"bri"`          // Full brightness
}

func LoadConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	// Set defaults
	if config.PollInterval == 0 {
		config.PollInterval = 30 * time.Second // 30 seconds default
	}

	return &config, nil
}
