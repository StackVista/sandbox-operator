package reaper

import (
	"context"
	"testing"
	"time"

	devopsv1 "github.com/stackvista/sandbox-operator/apis/devops/v1"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestShouldDetectExpiredSandboxUsingDefaultTTL(t *testing.T) {
	sandbox := newSandbox(time.Now().Add(-24*time.Hour), nil)
	assert.True(t, isExpired(context.Background(), sandbox, 1*time.Hour))
	assert.False(t, isExpired(context.Background(), sandbox, 25*time.Hour))
}
func TestShouldDetectExpiredSandboxUsingSandboxTTL(t *testing.T) {
	d := 1 * time.Hour
	sandbox := newSandbox(time.Now().Add(-2*time.Hour), &d)

	assert.True(t, isExpired(context.Background(), sandbox, 5*time.Hour))
}

func newSandbox(creation time.Time, ttl *time.Duration) devopsv1.Sandbox {
	return devopsv1.Sandbox{
		ObjectMeta: v1.ObjectMeta{
			CreationTimestamp: v1.NewTime(creation),
		},
		Spec: devopsv1.SandboxSpec{
			TTL: newDuration(ttl),
		},
	}
}

func newDuration(d *time.Duration) *v1.Duration {
	if d == nil {
		return nil
	}

	return &v1.Duration{
		Duration: *d,
	}
}
