package config

import (
	"os"

	"github.com/goccy/go-yaml"
)

type Channel struct {
	ID       string `yaml:"id"`
	Template string `yaml:"template"`
}

type ServerDetails struct {
	Channels []Channel `yaml:"channels"`
}

// type Role struct {
// 	ID    string `yaml:"id"`
// 	Level int    `yaml:"level"`
// }
//
// type RCONUsers struct {
// 	Roles []Role `yaml:"roles"`
// }

type Chat struct {
	ChannelID string `yaml:"channelID"`
}

type Config struct {
	ServerDetails *ServerDetails `yaml:"serverDetails,omitempty"`
	// RCONUsers     RCONUsers     `yaml:"rconUsers"`
	Chat *Chat `yaml:"chat,omitempty"`
}

func NewConfig(path string) (*Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
