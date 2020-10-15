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

func TestExpirationImminent(t *testing.T) {
	c := clk.NewMock()
	ctx := clock.WithContext(context.Background(), c)
	var tests = map[string]struct {
		createdAgo   time.Duration
		expiration   *time.Duration
		notification *time.Duration
		keepAlive    bool
		isImminent   bool
	}{
		"Default TTL: Not imminent if not hit first notify point":             {-2 * Day, nil, nil, false, false},
		"Default TTL: Imminent if hit first notify point":                     {-3 * Day, nil, nil, false, true},
		"Default TTL: Not imminent if keep alive":                             {-3 * Day, nil, nil, true, false},
		"Default TTL: Imminent if notified and not yet hit next notify point": {-4 * Day, nil, pDuration(-1 * Day), false, true},
		"Default TTL: Imminent if hit next notify point":                      {-5 * Day, nil, pDuration(-2 * Day), false, true},
		"Expiration date: Not imminent if not hit first notify point":         {-2 * Day, pDuration(5 * Day), nil, false, false},
		"Expiration date: Imminent if hit first notify point":                 {-2 * Day, pDuration(4 * Day), nil, false, true},
	}

	reaper := &Reaper{
		config: &Config{
			DefaultTtl:             7 * Day,
			FirstExpirationWarning: 4 * Day,
			WarningInterval:        2 * Day,
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			sandbox := newSandbox(c, data.createdAgo, data.expiration, data.keepAlive)
			sandbox.Status.LastNotification = newTime(pTime(c, data.notification))
			assert.Equal(t, reaper.isExpirationImminent(ctx, sandbox), data.isImminent)
		})
	}
}
func TestExpirationNotification(t *testing.T) {
	c := clk.NewMock()
	ctx := clock.WithContext(context.Background(), c)
	var tests = map[string]struct {
		createdAgo   time.Duration
		expiration   *time.Duration
		notification *time.Duration
		keepAlive    bool
		isNotify     bool
	}{
		"Default TTL: No notification if not hit first notify point":                 {-2 * Day, nil, nil, false, false},
		"Default TTL: Notification if hit first notify point":                        {-3 * Day, nil, nil, false, true},
		"Default TTL: Notification if keep alive and hit first notify point":         {-3 * Day, nil, nil, true, true},
		"Default TTL: No notification if notified and not yet hit next notify point": {-4 * Day, nil, pDuration(-1 * Day), false, false},
		"Default TTL: Notification if hit next notify point":                         {-5 * Day, nil, pDuration(-2 * Day), false, true},
		"Expiration date: No notification if not hit first notify point":             {-2 * Day, pDuration(5 * Day), nil, false, false},
		"Expiration date: Notification if hit first notify point":                    {-2 * Day, pDuration(4 * Day), nil, false, true},
	}

	reaper := &Reaper{
		config: &Config{
			DefaultTtl:             7 * Day,
			FirstExpirationWarning: 4 * Day,
			WarningInterval:        2 * Day,
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			sandbox := newSandbox(c, data.createdAgo, data.expiration, data.keepAlive)
			sandbox.Status.LastNotification = newTime(pTime(c, data.notification))
			assert.Equal(t, reaper.shouldNotify(ctx, sandbox), data.isNotify)
		})
	}
}

func TestConstructMessage(t *testing.T) {
	reaper := &Reaper{
		config: &Config{
			ReapMessage: "Sandbox `{{ .Sandbox.Name }}` is about to be deleted.",
		},
	}

	sb := devopsv1.Sandbox{
		ObjectMeta: v1.ObjectMeta{
			Name: "test-1",
		},
	}

	msg, err := reaper.constructMessage(context.Background(), reaper.config.ReapMessage, sb)
	assert.NilError(t, err)
	assert.Equal(t, msg, "Sandbox `test-1` is about to be deleted.")
}

func newSandbox(c clk.Clock, creation time.Duration, expiration *time.Duration, keepAlive bool) devopsv1.Sandbox {
	return devopsv1.Sandbox{
		ObjectMeta: v1.ObjectMeta{
			CreationTimestamp: v1.NewTime(c.Now().Add(creation)),
		},
		Spec: devopsv1.SandboxSpec{
			ExpirationDate: newTime(pTime(c, expiration)),
			ManualExpiry:   keepAlive,
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
