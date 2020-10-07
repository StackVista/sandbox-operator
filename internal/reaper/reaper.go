package reaper

import (
	"context"
	"fmt"
	"time"

	devopsv1 "gitlab.com/stackvista/devops/devopserator/apis/devops/v1"
	"gitlab.com/stackvista/devops/devopserator/pkg/client/versioned"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caspr-io/mu-kit/kubernetes"
	"github.com/slack-go/slack"
	"gitlab.com/stackvista/devops/devopserator/internal/logr"
)

type Config struct {
	DefaultTtl         time.Duration `split_words:"true" required:"true" default:"604800s"` // Default 1 week
	SlackApiKey        string        `split_words:"true" required:"true"`
	SlackChannelID     string        `split_words:"true" required:"true"`
	SlackPostAsUser    string        `split_words:"true" required:"false"`
	SlackPostAsIconURL string        `split_words:"true" required:"false"`
	SlackPostMessage   string        `split_words:"true" required:"true"`
}

func ReapNamespaces(ctx context.Context, config *Config) error {
	logger := logr.Ctx(ctx)
	k8s, err := kubernetes.ConnectToKubernetes()
	if err != nil {
		return err
	}

	logger.Info("Connected to Kubernetes")

	sandboxes, err := listSandboxes(ctx, k8s)
	if err != nil {
		return err
	}

	for _, sb := range sandboxes {
		if isExpired(ctx, sb, config.DefaultTtl) {
			if err := deleteSandbox(ctx, k8s, sb); err != nil {
				return err
			}

			if err := notifyDeletion(ctx, config, sb); err != nil {
				return err
			}
		}
	}
	return nil
}

func deleteSandbox(ctx context.Context, k8s *kubernetes.K8s, sb devopsv1.Sandbox) error {
	sandboxClient := versioned.New(k8s.Clientset.RESTClient())

	return sandboxClient.DevopsV1().Sandboxes().Delete(ctx, sb.Name, v1.DeleteOptions{})
}

func listSandboxes(ctx context.Context, k8s *kubernetes.K8s) ([]devopsv1.Sandbox, error) {
	sandboxClient := versioned.New(k8s.Clientset.RESTClient())

	sbList, err := sandboxClient.DevopsV1().Sandboxes().List(ctx, v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return sbList.Items, nil
}

func isExpired(ctx context.Context, sb devopsv1.Sandbox, defaultTTL time.Duration) bool {
	ttl := defaultTTL
	if sb.Spec.TTL != nil {
		ttl = sb.Spec.TTL.Duration
	}

	return sb.CreationTimestamp.Add(ttl).Before(time.Now())
}

func notifyDeletion(ctx context.Context, config *Config, sb devopsv1.Sandbox) error {
	client := slack.New(config.SlackApiKey)

	msg := fmt.Sprintf("<@%s>, "+config.SlackPostMessage, sb.Spec.SlackId, sb.Name)

	msgOpts := []slack.MsgOption{
		slack.MsgOptionText(msg, false),
	}

	if config.SlackPostAsUser != "" {
		msgOpts = append(msgOpts, slack.MsgOptionUsername(config.SlackPostAsUser))
	}

	if config.SlackPostAsIconURL != "" {
		msgOpts = append(msgOpts, slack.MsgOptionIconURL(config.SlackPostAsIconURL))
	}

	if _, _, err := client.PostMessage(config.SlackChannelID, msgOpts...); err != nil {
		return err
	}

	return nil
}
