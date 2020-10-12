package reaper

import (
	"context"
	"testing"
	"time"

	clk "github.com/benbjohnson/clock"
	devopsv1 "github.com/stackvista/sandbox-operator/apis/devops/v1"
	"github.com/stackvista/sandbox-operator/internal/clock"
	"gotest.tools/v3/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var Day = 24 * time.Hour

func TestExpiration(t *testing.T) {
	c := clk.NewMock()
	ctx := clock.WithContext(context.Background(), c)

	var tests = map[string]struct {
		createdAgo time.Duration
		expiration *time.Duration
		keepAlive  bool
		isExpired  bool
	}{
		"Expired by default TTL":                             {-3 * Day, nil, false, true},
		"Not yet expired by default TTL":                     {-1 * Day, nil, false, false},
		"Not yet expired by default TTL as KeepAlive is set": {-3 * Day, nil, true, false},
		"Expired by expiration date":                         {-3 * Day, pDuration(-1 * Day), false, true},
		"Not yet expired by expiration date":                 {-3 * Day, pDuration(1 * Day), false, false},
		"Not expired by expiration date as KeepAlive is set": {-3 * Day, pDuration(-1 * Day), true, false},
	}

	reaper := &Reaper{
		config: &Config{
			DefaultTtl: 2 * Day,
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			sandbox := newSandbox(c, data.createdAgo, data.expiration, data.keepAlive)
			assert.Equal(t, reaper.isExpired(ctx, sandbox), data.isExpired)
		})
	}
}

func TestNotificationImminent(t *testing.T) {
	c := clk.NewMock()
	ctx := clock.WithContext(context.Background(), c)
	var tests = map[string]struct {
		createdAgo   time.Duration
		expiration   *time.Duration
		notification *time.Duration
		keepAlive    bool
		isNotify     bool
	}{
		"Default TTL: No warning if not hit first notify point":                 {-2 * Day, nil, nil, false, false},
		"Default TTL: Warning if hit first notify point":                        {-3 * Day, nil, nil, false, true},
		"Default TTL: No warning if keep alive":                                 {-3 * Day, nil, nil, true, false},
		"Default TTL: No warning if notified and not yet hit next notify point": {-4 * Day, nil, pDuration(-1 * Day), false, false},
		"Default TTL: Warning if hit next notify point":                         {-5 * Day, nil, pDuration(-2 * Day), false, true},
		"Expiration date: No warning if not hit first notify point":             {-2 * Day, pDuration(5 * Day), nil, false, false},
		"Expiration date: Warning if hit first notify point":                    {-2 * Day, pDuration(4 * Day), nil, false, true},
	}

	reaper := &Reaper{
		config: &Config{
			DefaultTtl:             7 * Day,
			FirstExpirationWarning: 4 * Day,
			NotificationInterval:   2 * Day,
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			sandbox := newSandbox(c, data.createdAgo, data.expiration, data.keepAlive)
			sandbox.Status.LastNotification = newTime(pTime(c, data.notification))
			assert.Equal(t, reaper.isExpirationImminent(ctx, sandbox), data.isNotify)
		})
	}
}

func newSandbox(c clk.Clock, creation time.Duration, expiration *time.Duration, keepAlive bool) devopsv1.Sandbox {
	return devopsv1.Sandbox{
		ObjectMeta: v1.ObjectMeta{
			CreationTimestamp: v1.NewTime(c.Now().Add(creation)),
		},
		Spec: devopsv1.SandboxSpec{
			ExpirationDate: newTime(pTime(c, expiration)),
			KeepAlive:      keepAlive,
		},
	}
}

func pDuration(d time.Duration) *time.Duration {
	return &d
}

func pTime(c clk.Clock, d *time.Duration) *time.Time {
	if d == nil {
		return nil
	}

	t := c.Now().Add(*d)
	return &t
}

func newTime(t *time.Time) *v1.Time {
	if t == nil {
		return nil
	}

	return &v1.Time{
		Time: *t,
	}
}
