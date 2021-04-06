package kubernetes

import (
	home "github.com/mitchellh/go-homedir"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func LoadConfig() (*rest.Config, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil && err != rest.ErrNotInCluster {
		return nil, err
	} else if err != nil {
		kubeconfig, err := home.Expand("~/.kube/config")
		if err != nil {
			return nil, err
		}
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}

	return cfg, nil
}
