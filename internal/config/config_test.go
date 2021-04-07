package config

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestShouldUnmarshalClientConfig(t *testing.T) {
	config := `
scaler:
  system-namespaces:
    - kube-system
    - logging
`

	d := yaml.NewDecoder(bytes.NewBufferString(config))
	cfg := &Config{}
	assert.NoError(t, d.Decode(&cfg))
	assert.Contains(t, cfg.Scaler.SystemNamespaces, "kube-system")
	assert.Contains(t, cfg.Scaler.SystemNamespaces, "logging")
	assert.Equal(t, 4*time.Hour, cfg.Scaler.ScaleInterval)
}
