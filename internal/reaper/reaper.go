package reaper

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	devopsv1 "github.com/stackvista/sandbox-operator/apis/devops/v1"
	"github.com/stackvista/sandbox-operator/pkg/client/versioned"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caspr-io/mu-kit/kubernetes"
	"github.com/stackvista/sandbox-operator/internal/slack"
)

type Config struct {
	DefaultTtl  time.Duration `split_words:"true" required:"true" default:"604800s"` // Default 1 week
	ReapMessage string        `split_words:"true" required:"true"`
	Slack       *slack.Config `split_words:"true" required:"true"`
}

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

		if isExpired(ctx, sb, r.config.DefaultTtl) {
			logger.Info().Str("sandbox", sb.Name).Msg("Sandbox is expired.")

			return r.deleteAndNotify(ctx, sb)
		} // TODO almost expired, notify once per day
	}

	return nil
}

func (r *Reaper) deleteAndNotify(ctx context.Context, sb devopsv1.Sandbox) error {
	if err := r.sandboxClient.DevopsV1().Sandboxes().Delete(ctx, sb.Name, v1.DeleteOptions{}); err != nil {
		return err
	}

	return slack.NewSlacker(r.config.Slack).NotifyUser(sb.Spec.SlackId, "", r.config.ReapMessage)
}

// isExpired checks whether the given Sandbox has expired its TTL
func isExpired(ctx context.Context, sb devopsv1.Sandbox, defaultTTL time.Duration) bool {
	ttl := defaultTTL
	if sb.Spec.TTL != nil {
		ttl = sb.Spec.TTL.Duration
	}

	return sb.CreationTimestamp.Add(ttl).Before(time.Now())
}
