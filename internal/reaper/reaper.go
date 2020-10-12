package reaper

import (
	"context"
	"time"

	"github.com/stackvista/sandbox-operator/internal/clock"

	"github.com/rs/zerolog/log"
	devopsv1 "github.com/stackvista/sandbox-operator/apis/devops/v1"
	"github.com/stackvista/sandbox-operator/pkg/client/versioned"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caspr-io/mu-kit/kubernetes"
	"github.com/stackvista/sandbox-operator/internal/slack"
)

type Config struct {
	DefaultTtl               time.Duration `split_words:"true" required:"true" default:"604800s"` // Default 1 week
	FirstExpirationWarning   time.Duration `split_words:"true" required:"true" default:"259200s"` // Default 3 days
	NotificationInterval     time.Duration `split_words:"true" required:"true" default:"86400s"`  // Default 1 day
	ExpirationWarningMessage string        `split_words:"true" required:"true"`
	ReapMessage              string        `split_words:"true" required:"true"`
	ExpirationOverdueMessage string        `split_words:"true" required:"true"`
	Slack                    *slack.Config `split_words:"true" required:"true"`
}

// Reaper will reap sandboxes from the cluster.
type Reaper struct {
	sandboxClient *versioned.Clientset
	config        *Config
}

func NewReaper(ctx context.Context, config *Config) (*Reaper, error) {
	logger := log.Ctx(ctx)
	k8s, err := kubernetes.ConnectToKubernetes()
	if err != nil {
		return nil, err
	}
	logger.Info().Msg("Connected to Kubernetes")

	return &Reaper{
		sandboxClient: versioned.New(k8s.Clientset.RESTClient()),
		config:        config,
	}, nil
}

func (r *Reaper) Run(ctx context.Context) error {
	logger := log.Ctx(ctx)

	sandboxes, err := r.sandboxClient.DevopsV1().Sandboxes().List(ctx, v1.ListOptions{})
	if err != nil {
		return err
	}

	for _, sb := range sandboxes.Items {
		logger.Debug().Str("sandbox", sb.Name).Msg("Inspecting Sandbox")

		if r.isExpired(ctx, sb) {
			logger.Info().Str("sandbox", sb.Name).Msg("Sandbox is expired.")

			return r.deleteAndNotify(ctx, sb)
		} else if r.isExpirationImminent(ctx, sb) {
			logger.Info().Str("sandbox", sb.Name).Msg("Warning about imminent expiration")
			// return r.notifyExpirationImminent(ctx, sb)
		}
	}

	logger.Info().Msg("Finished reaping run")

	return nil
}

func (r *Reaper) deleteAndNotify(ctx context.Context, sb devopsv1.Sandbox) error {
	if err := r.sandboxClient.DevopsV1().Sandboxes().Delete(ctx, sb.Name, v1.DeleteOptions{}); err != nil {
		return err
	}

	return slack.NewSlacker(r.config.Slack).NotifyUser(sb.Spec.SlackId, "", r.config.ReapMessage)
}

func (r *Reaper) expirationDate(ctx context.Context, sb devopsv1.Sandbox) *time.Time {
	if sb.Spec.KeepAlive {
		return nil // No expiration date if KeepAlive is set
	}

	if sb.Spec.ExpirationDate != nil {
		return &sb.Spec.ExpirationDate.Time
	} else {
		t := sb.CreationTimestamp.Add(r.config.DefaultTtl)
		return &t
	}
}

// isExpired checks whether the given Sandbox has expired its TTL
func (r *Reaper) isExpired(ctx context.Context, sb devopsv1.Sandbox) bool {
	expDate := r.expirationDate(ctx, sb)
	if expDate == nil {
		return false
	}

	return clock.Ctx(ctx).Now().After(*expDate)
}

// isExpirationImminent checks whether the user should be notified that expiration of the sandbox is imminent
func (r *Reaper) isExpirationImminent(ctx context.Context, sb devopsv1.Sandbox) bool {
	expDate := r.expirationDate(ctx, sb)
	if expDate == nil {
		return false
	}

	now := clock.Ctx(ctx).Now()
	if sb.Status.LastNotification == nil {
		firstNotify := expDate.Add(-r.config.FirstExpirationWarning)
		return now.After(firstNotify) || now.Equal(firstNotify)
	}

	nextNotify := sb.Status.LastNotification.Add(r.config.NotificationInterval)
	return now.After(nextNotify) || now.Equal(nextNotify)
}
