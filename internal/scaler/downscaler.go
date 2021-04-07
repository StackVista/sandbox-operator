package scaler

import (
	"context"
	"crypto/sha512"
	"net/http"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"

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
	ConfigMapKey = "com.stackstate.devops.checksum"
)

type DownScaler struct {
	klient   client.Client
	config   *config.ScalerConfig
	notifier notification.Notifier
}

type ChangeChecksum struct {
	Checksum string
	Date     time.Time
}

func NewScaler(ctx context.Context, config *config.ScalerConfig, notifier notification.Notifier) (*DownScaler, error) {
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

		cs, err := d.GetPreviousChecksum(ctx, ns)
		if err != nil {
			return err
		}

		log.Ctx(ctx).Info().Str("namespace", ns.Name).Interface("checksum", cs).Msg("Inspecting namespace")

	}

	return nil
}

func (d *DownScaler) CalculateChangeChecksum(ctx context.Context, ns corev1.Namespace) (string, error) {
	shaSum := sha512.New()

	for _, f := range []K8sChecksum{ChecksumDeployments, ChecksumDaemonSets, ChecksumReplicaSets, ChecksumStatefulSets} {
		if err := f(ctx, d.klient, shaSum); err != nil {
			return "", err
		}
	}

	return string(shaSum.Sum(nil)), nil
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

	date, err := time.Parse(time.RFC3339, cm.Data["date"])
	if err != nil {
		return nil, err
	}

	return &ChangeChecksum{
		Checksum: cm.Data["checksum"],
		Date:     date,
	}, nil
}
