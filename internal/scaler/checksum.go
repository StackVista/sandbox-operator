package scaler

import (
	"context"
	"crypto/sha512"
	"encoding/base64"
	"hash"

	"github.com/rs/zerolog/log"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// K8sChecksum is a function which calculates a K8s Checksum
type K8sChecksum func(ctx context.Context, c client.Client, ns corev1.Namespace, h hash.Hash) error

func CalculateChangeChecksum(ctx context.Context, c client.Client, ns corev1.Namespace) (string, error) {
	shaSum := sha512.New()

	for _, f := range []K8sChecksum{ChecksumDeployments, ChecksumDaemonSets, ChecksumReplicaSets, ChecksumStatefulSets} {
		if err := f(ctx, c, ns, shaSum); err != nil {
			return "", err
		}
	}

	sum := shaSum.Sum(nil)
	return base64.StdEncoding.EncodeToString(sum), nil
}

// ChecksumDeployments calculates the checksum of the Generation of all Deployment objects
func ChecksumDeployments(ctx context.Context, c client.Client, ns corev1.Namespace, h hash.Hash) error {
	return WithDeployments(ctx, c, ns, func(d appsv1.Deployment) error {
		return hashResource(ctx, h, d.Name, d.Kind, d.Generation)
	})
}

// ChecksumReplicaSets calculates the checksum of the Generation of all ReplicaSet objects
func ChecksumReplicaSets(ctx context.Context, c client.Client, ns corev1.Namespace, h hash.Hash) error {
	return WithReplicaSets(ctx, c, ns, func(d appsv1.ReplicaSet) error {
		return hashResource(ctx, h, d.Name, d.Kind, d.Generation)
	})
}

// ChecksumDaemonSets calculates the checksum of the Generation of all DaemonSet objects
func ChecksumDaemonSets(ctx context.Context, c client.Client, ns corev1.Namespace, h hash.Hash) error {
	return WithDaemonSets(ctx, c, ns, func(d appsv1.DaemonSet) error {
		return hashResource(ctx, h, d.Name, d.Kind, d.Generation)
	})
}

// ChecksumStatefulSets calculates the checksum of the Generation of all StatefulSet objects
func ChecksumStatefulSets(ctx context.Context, c client.Client, ns corev1.Namespace, h hash.Hash) error {
	return WithStatefulSets(ctx, c, ns, func(d appsv1.StatefulSet) error {
		return hashResource(ctx, h, d.Name, d.Kind, d.Generation)
	})
}

func hashResource(ctx context.Context, h hash.Hash, name, kind string, generation int64) error {
	log.Ctx(ctx).Trace().Int64("generation", generation).Str("name", name).Str("kind", kind).Msg("Adding generation to checksum")
	if _, err := h.Write([]byte(name)); err != nil {
		return err
	}

	if _, err := h.Write(int64AsBytes(generation)); err != nil {
		return err
	}

	return nil
}

func int64AsBytes(i int64) []byte {
	return []byte{
		byte(0xff & i),
		byte(0xff & (i >> 8)),
		byte(0xff & (i >> 16)),
		byte(0xff & (i >> 24)),
		byte(0xff & (i >> 32)),
		byte(0xff & (i >> 40)),
		byte(0xff & (i >> 48)),
		byte(0xff & (i >> 56)),
	}
}
