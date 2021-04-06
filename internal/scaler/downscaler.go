package scaler

import (
	"context"

	corev1 "k8s.io/api/core/v1"

	"github.com/stackvista/sandbox-operator/internal/config"
	"github.com/stackvista/sandbox-operator/internal/notification"
	"github.com/stackvista/sandbox-operator/pkg/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DownScaler struct {
	klient   client.Client
	config   *config.ReaperConfig
	notifier notification.Notifier
}

func NewScaler(ctx context.Context, config *config.ReaperConfig, notifier notification.Notifier) (*DownScaler, error) {
	cfg, err := kubernetes.LoadConfig()
	if err != nil {
		return nil, err
	}

	c, err := client.New(cfg, client.Options{})
	if err != nil {
		return nil, err
	}

	return &DownScaler{
		klient:   c,
		config:   config,
		notifier: notifier,
	}, nil
}

func (d *DownScaler) Run(ctx context.Context) error {
	nsList := &corev1.NamespaceList{}
	if err := d.klient.List(ctx, nsList); err != nil {
		return err
	}

	for _, ns := range nsList.Items {
		println(ns.Name)
	}

	return nil
}
