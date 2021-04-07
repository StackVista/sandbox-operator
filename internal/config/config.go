package config

import (
	"time"

	"github.com/mcuadros/go-defaults"
)

type Config struct {
	Slack  SlackConfig  `yaml:"slack"`
	Reaper ReaperConfig `yaml:"reaper"`
	Scaler ScalerConfig `yaml:"scaler"`
}

type ReaperConfig struct {
	DefaultTtl               time.Duration `yaml:"default-ttl" default:"168h"`             // Default 1 week
	FirstExpirationWarning   time.Duration `yaml:"first-expiration-warning" default:"72h"` // Default 3 days
	WarningInterval          time.Duration `yaml:"warning-interval" default:"24h"`         // Default 1 day
	ExpirationWarningMessage string        `yaml:"expiration-warning-message"`
	ReapMessage              string        `yaml:"reap-message"`
	ExpirationOverdueMessage string        `yaml:"expiration-overdue-message"`
}

type SlackConfig struct {
	ApiKey        string `yaml:"api-key"`
	ChannelID     string `yaml:"channel-id"`
	PostAsUser    string `yaml:"post-as-user"`
	PostAsIconURL string `yaml:"post-as-icon-url"`
}

type ScalerConfig struct {
	ScaleInterval    time.Duration `yaml:"scale-interval" default:"4h"`
	SystemNamespaces []string      `yaml:"system-namespaces" default:"kube-system,kube-public,kube-node-lease"`
}

func (s *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	defaults.SetDefaults(s)

	type cfg Config

	if err := unmarshal((*cfg)(s)); err != nil {
		return err
	}

	return nil
}
