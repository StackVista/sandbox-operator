package scaler

import (
	"context"
	"hash"

	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// K8sChecksum is a function which calculates a K8s Checksum
type K8sChecksum func(ctx context.Context, c client.Client, h hash.Hash) error

// ChecksumDeployments calculates the checksum of the Generation of all Deployment objects
func ChecksumDeployments(ctx context.Context, c client.Client, h hash.Hash) error {
	l := &appsv1.DeploymentList{}

	if err := c.List(ctx, l); err != nil {
		return err
	}

	for _, d := range l.Items {
		_, err := h.Write(int64AsBytes(d.Generation))
		if err != nil {
			return err
		}
	}

	return nil
}

// ChecksumReplicaSets calculates the checksum of the Generation of all ReplicaSet objects
func ChecksumReplicaSets(ctx context.Context, c client.Client, h hash.Hash) error {
	l := &appsv1.ReplicaSetList{}

	if err := c.List(ctx, l); err != nil {
		return err
	}

	for _, d := range l.Items {
		_, err := h.Write(int64AsBytes(d.Generation))
		if err != nil {
			return err
		}
	}

	return nil
}

// ChecksumDaemonSets calculates the checksum of the Generation of all DaemonSet objects
func ChecksumDaemonSets(ctx context.Context, c client.Client, h hash.Hash) error {
	l := &appsv1.DaemonSetList{}

	if err := c.List(ctx, l); err != nil {
		return err
	}

	for _, d := range l.Items {
		_, err := h.Write(int64AsBytes(d.Generation))
		if err != nil {
			return err
		}
	}

	return nil
}

// ChecksumStatefulSets calculates the checksum of the Generation of all StatefulSet objects
func ChecksumStatefulSets(ctx context.Context, c client.Client, h hash.Hash) error {
	l := &appsv1.StatefulSetList{}

	if err := c.List(ctx, l); err != nil {
		return err
	}

	for _, d := range l.Items {
		_, err := h.Write(int64AsBytes(d.Generation))
		if err != nil {
			return err
		}
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
