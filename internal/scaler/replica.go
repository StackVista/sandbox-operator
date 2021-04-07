package scaler

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ReplicaCount map[string]int

type FetchReplicas func(ctx context.Context, c client.Client, ns corev1.Namespace) (ReplicaCount, error)

func RetrieveReplicas(ctx context.Context, c client.Client, ns corev1.Namespace) (map[string]ReplicaCount, error) {
	m := map[string]ReplicaCount{}

	funcs := map[string]FetchReplicas{
		"deployment":  FetchDeploymentReplicas,
		"replicaset":  FetchReplicaSetReplicas,
		"statefulset": FetchStatefulSetReplicas,
	}

	for t, f := range funcs {
		r, err := f(ctx, c, ns)
		if err != nil {
			return nil, err
		}

		m[t] = r
	}
	return m, nil
}

func FetchDeploymentReplicas(ctx context.Context, c client.Client, ns corev1.Namespace) (ReplicaCount, error) {
	r := ReplicaCount{}
	err := WithDeployments(ctx, c, ns, func(d appsv1.Deployment) error {
		reps := d.Spec.Replicas
		if reps == nil {
			r[d.Name] = 1
		} else {
			r[d.Name] = int(*reps)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return r, nil
}

func FetchStatefulSetReplicas(ctx context.Context, c client.Client, ns corev1.Namespace) (ReplicaCount, error) {
	r := ReplicaCount{}
	err := WithStatefulSets(ctx, c, ns, func(d appsv1.StatefulSet) error {
		reps := d.Spec.Replicas
		if reps == nil {
			r[d.Name] = 1
		} else {
			r[d.Name] = int(*reps)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return r, nil
}

func FetchReplicaSetReplicas(ctx context.Context, c client.Client, ns corev1.Namespace) (ReplicaCount, error) {
	r := ReplicaCount{}
	err := WithReplicaSets(ctx, c, ns, func(d appsv1.ReplicaSet) error {
		reps := d.Spec.Replicas
		if reps == nil {
			r[d.Name] = 1
		} else {
			r[d.Name] = int(*reps)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return r, nil

}
