package client

import (
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// SyncConfig contains the configuration options for the SyncSecrets operation.
type SyncConfig struct {
	PodNamespace string

	LivenessPort  int
	ResyncPeriod  time.Duration
	EnableTraefik bool

	OutOfCluster bool
	KubeConfig   string
}

func InitializeClient(config *SyncConfig) (*kubernetes.Clientset, error) {
	restConfig, err := InitializeClusterConfig(config)
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(restConfig)
}

func InitializeClusterConfig(config *SyncConfig) (*rest.Config, error) {
	if config.OutOfCluster {
		return clientcmd.BuildConfigFromFlags("", config.KubeConfig)
	}

	return rest.InClusterConfig()
}
