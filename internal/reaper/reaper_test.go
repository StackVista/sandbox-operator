package reaper

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	devopsv1 "gitlab.com/stackvista/devops/devopserator/apis/devops/v1"
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
