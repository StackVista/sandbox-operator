package config

import (
	"time"

	"github.com/creasty/defaults"
)

type Config struct {
	Slack  *SlackConfig  `yaml:"slack"`
	Reaper *ReaperConfig `yaml:"reaper"`
	Scaler *ScalerConfig `yaml:"scaler"`
}

type ReaperConfig struct {
	DefaultTtl               time.Duration `yaml:"defaultTtl" default:"168h"`            // Default 1 week
	FirstExpirationWarning   time.Duration `yaml:"firstExpirationWarning" default:"72h"` // Default 3 days
	WarningInterval          time.Duration `yaml:"warningInterval" default:"24h"`        // Default 1 day
	ExpirationWarningMessage string        `yaml:"expirationWarningMessage"`
	ReapMessage              string        `yaml:"reapMessage"`
	ExpirationOverdueMessage string        `yaml:"expirationOverdueMessage"`
}

type SlackConfig struct {
	ApiKey        string `yaml:"apiKey"`
	ChannelID     string `yaml:"channelID"`
	PostAsUser    string `yaml:"postAsUser"`
	PostAsIconURL string `yaml:"postAsIconURL"`
}

type ScalerConfig struct {
}

func (s *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := defaults.Set(s); err != nil {
		return err
	}

	type cfg Config

	if err := unmarshal((*cfg)(s)); err != nil {
		return err
	}

	return nil
}
