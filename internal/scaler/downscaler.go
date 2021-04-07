package scaler

import (
	"context"
	"net/http"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/rs/zerolog/log"
	"github.com/stackvista/sandbox-operator/internal/config"
	"github.com/stackvista/sandbox-operator/internal/notification"
	"github.com/stackvista/sandbox-operator/pkg/kubernetes"
	"github.com/stackvista/sandbox-operator/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ConfigMapKey    = "com.stackstate.devops.checksum"
	ChecksumDataKey = "checksum"
	DateDataKey     = "date"
)

type DownScaler struct {
	klient   client.Client
	config   config.ScalerConfig
	notifier notification.Notifier
}

type ChangeChecksum struct {
	Checksum string
	Date     time.Time
}

func NewScaler(ctx context.Context, config config.ScalerConfig, notifier notification.Notifier) (*DownScaler, error) {
	cfg, err := kubernetes.LoadConfig()
	if err != nil {
		return nil, err
	}

	c, err := client.New(cfg, client.Options{})
	if err != nil {
		return nil, err
	}

	log.Info().Interface("config", config).Msg("Starting scaler")

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
		if util.ContainsString(d.config.SystemNamespaces, ns.Name) {
			log.Ctx(ctx).Debug().Str("namespace", ns.Name).Msg("Skipping system namespace")
			continue
		}

		if err := d.handleNamespace(ctx, ns); err != nil {
			return err
		}
	}

	return nil
}

func (d *DownScaler) handleNamespace(ctx context.Context, ns corev1.Namespace) error {
	logger := log.Ctx(ctx).With().Str("namespace", ns.Name).Logger()

	logger.Info().Msg("Inspecting namespace")

	cs, err := d.GetPreviousChecksum(ctx, ns)
	if err != nil {
		return err
	}

	newChecksum, err := CalculateChangeChecksum(ctx, d.klient, ns)
	if err != nil {
		return err
	}

	if newChecksum != cs.Checksum {
		logger.Info().Str("new-checksum", newChecksum).Str("old-checksum", cs.Checksum).Msg("Detected checksum change")
		return d.StoreChecksum(ctx, ns, newChecksum)
	}

	if cs.Date.Add(d.config.ScaleInterval).After(time.Now()) {
		logger.Info().Time("last-modified", cs.Date).Dur("scale-interval", d.config.ScaleInterval).Msg("Scale down time not reached")
		return nil
	}

	return nil
}

func (d *DownScaler) GetPreviousChecksum(ctx context.Context, ns corev1.Namespace) (*ChangeChecksum, error) {
	cm := &corev1.ConfigMap{}
	if err := d.klient.Get(ctx, types.NamespacedName{Namespace: ns.Name, Name: ConfigMapKey}, cm); err != nil {
		if serr, ok := err.(*errors.StatusError); ok {
			if serr.ErrStatus.Code == http.StatusNotFound {
				return &ChangeChecksum{
					Checksum: "",
					Date:     time.Now(),
				}, nil
			}
		}

		return nil, err
	}

	date, err := time.Parse(time.RFC3339, cm.Data[DateDataKey])
	if err != nil {
		return nil, err
	}

	return &ChangeChecksum{
		Checksum: cm.Data[ChecksumDataKey],
		Date:     date,
	}, nil
}

func (d *DownScaler) StoreChecksum(ctx context.Context, ns corev1.Namespace, checksum string) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		cm := &corev1.ConfigMap{}
		if err := d.klient.Get(ctx, types.NamespacedName{Namespace: ns.Name, Name: ConfigMapKey}, cm); err != nil {
			if serr, ok := err.(*errors.StatusError); ok {
				if serr.ErrStatus.Code == http.StatusNotFound {
					return d.klient.Create(ctx, &corev1.ConfigMap{
						ObjectMeta: v1.ObjectMeta{
							Namespace: ns.Name,
							Name:      ConfigMapKey,
						},
						Data: map[string]string{
							ChecksumDataKey: checksum,
							DateDataKey:     time.Now().Format(time.RFC3339),
						},
					})
				}
			}

			return err
		}

		cm.Data[ChecksumDataKey] = checksum
		cm.Data[DateDataKey] = time.Now().Format(time.RFC3339)

		return d.klient.Update(ctx, cm)
	})
}
