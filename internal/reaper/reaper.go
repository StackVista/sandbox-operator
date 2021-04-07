package reaper

import (
	"bytes"
	"context"
	"text/template"
	"time"

	"github.com/stackvista/sandbox-operator/internal/clock"
	"github.com/stackvista/sandbox-operator/internal/config"
	"github.com/stackvista/sandbox-operator/internal/notification"

	"github.com/rs/zerolog/log"
	devopsv1 "github.com/stackvista/sandbox-operator/apis/devops/v1"
	"github.com/stackvista/sandbox-operator/pkg/client/versioned"
	"github.com/stackvista/sandbox-operator/pkg/kubernetes"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Reaper will reap sandboxes from the cluster.
type Reaper struct {
	sandboxClient *versioned.Clientset
	config        config.ReaperConfig
	notifier      notification.Notifier
}

func NewReaper(ctx context.Context, config config.ReaperConfig, notifier notification.Notifier) (*Reaper, error) {
	logger := log.Ctx(ctx)

	cfg, err := kubernetes.LoadConfig()
	if err != nil {
		return nil, err
	}

	client, err := versioned.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	logger.Info().Msg("Connected to Kubernetes")

	return &Reaper{
		sandboxClient: client,
		config:        config,
		notifier:      notifier,
	}, nil
}

func (r *Reaper) Run(ctx context.Context) error {
	logger := log.Ctx(ctx)

	logger.Info().Msg("Going to list sandboxes...")

	sandboxes, err := r.sandboxClient.DevopsV1().Sandboxes().List(ctx, v1.ListOptions{})
	if err != nil {
		logger.Error().Err(err).Msg("Error while listing sandboxes")
		return err
	}

	for _, sb := range sandboxes.Items {
		logger.Debug().Str("sandbox", sb.Name).Msg("Inspecting Sandbox")

		if r.isExpired(ctx, sb) {
			logger.Info().Str("sandbox", sb.Name).Msg("Sandbox is expired.")

			if err := r.sandboxClient.DevopsV1().Sandboxes().Delete(ctx, sb.Name, v1.DeleteOptions{}); err != nil {
				return err
			}

			if err := r.notify(ctx, r.config.ReapMessage, sb); err != nil {
				return err
			}

		} else if r.isExpirationImminent(ctx, sb) {
			if r.shouldNotify(ctx, sb) {
				logger.Info().Str("sandbox", sb.Name).Msg("Warning about imminent expiration")

				if err := r.notify(ctx, r.config.ExpirationWarningMessage, sb); err != nil {
					return err
				}

				if err := r.updateLastNotificationDate(ctx, sb); err != nil {
					return err
				}
			}
		} else if r.isExpirationOverdue(ctx, sb) {
			if r.shouldNotify(ctx, sb) {
				logger.Info().Str("sandbox", sb.Name).Msg("Manual expiry sandbox is overdue, notifying user")

				if err := r.notify(ctx, r.config.ExpirationOverdueMessage, sb); err != nil {
					return err
				}

				if err := r.updateLastNotificationDate(ctx, sb); err != nil {
					return err
				}
			}

		}
	}

	logger.Info().Msg("Finished reaping run")

	return nil
}

// notify uses the Reaper.notifier to notify that the Sandbox is either reaped, or will be reaped.
func (r *Reaper) notify(ctx context.Context, message string, sb devopsv1.Sandbox) error {
	msg, err := r.constructMessage(ctx, message, sb)
	if err != nil {
		return err
	}

	return r.notifier.Notify("", msg)
}

// updateLastNotificationDate updates the Sandbox.Status.LastNotification field with the date of `now`.
func (r *Reaper) updateLastNotificationDate(ctx context.Context, sb devopsv1.Sandbox) error {
	sb.Status.LastNotification = &v1.Time{Time: clock.Ctx(ctx).Now()}

	if _, err := r.sandboxClient.DevopsV1().Sandboxes().UpdateStatus(ctx, &sb, v1.UpdateOptions{}); err != nil {
		return err
	}

	return nil
}

func (r *Reaper) expirationDate(ctx context.Context, sb devopsv1.Sandbox) time.Time {
	if sb.Spec.ExpirationDate != nil {
		return sb.Spec.ExpirationDate.Time
	} else {
		return sb.CreationTimestamp.Add(r.config.DefaultTtl)
	}
}

// isExpired checks whether the given Sandbox has expired its TTL
func (r *Reaper) isExpired(ctx context.Context, sb devopsv1.Sandbox) bool {
	if sb.Spec.ManualExpiry {
		return false // No expiration if KeepAlive is set
	}

	expDate := r.expirationDate(ctx, sb)

	return clock.Ctx(ctx).Now().After(expDate)
}

// isExpirationImminent checks whether the Sandbox will soon be reaped
func (r *Reaper) isExpirationImminent(ctx context.Context, sb devopsv1.Sandbox) bool {
	if sb.Spec.ManualExpiry {
		return false // No expiration if KeepAlive is set
	}

	expDate := r.expirationDate(ctx, sb)
	now := clock.Ctx(ctx).Now()
	warnDate := expDate.Add(-r.config.FirstExpirationWarning)

	return now.After(warnDate) || now.Equal(warnDate)
}

// shouldNotify checks whether the Sandbox owner should be notified about a pending expiry
func (r *Reaper) shouldNotify(ctx context.Context, sb devopsv1.Sandbox) bool {
	expDate := r.expirationDate(ctx, sb)
	now := clock.Ctx(ctx).Now()

	notification := expDate.Add(-r.config.FirstExpirationWarning)
	if sb.Status.LastNotification != nil {
		notification = sb.Status.LastNotification.Add(r.config.WarningInterval)
	}

	return now.After(notification) || now.Equal(notification)

}

// isExpirationOverdue checks whether a Sandbox that has Sandbox.Spec.KeepAlive set has passed its expiry.
func (r *Reaper) isExpirationOverdue(ctx context.Context, sb devopsv1.Sandbox) bool {
	if !sb.Spec.ManualExpiry {
		return false // If not KeepAlive, it is not overdue
	}

	expDate := r.expirationDate(ctx, sb)
	now := clock.Ctx(ctx).Now()

	return now.After(expDate) || now.Equal(expDate)
}

// constructMessage templates the configured message using the Sandbox as context.
func (r *Reaper) constructMessage(ctx context.Context, message string, sb devopsv1.Sandbox) (string, error) {
	m := map[string]interface{}{
		"Sandbox":        sb,
		"ExpirationDate": r.expirationDate(ctx, sb),
	}

	t, err := template.New("notification").Parse(message)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, m); err != nil {
		return "", err
	}

	return buf.String(), nil
}
