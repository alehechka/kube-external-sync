package client

// SyncConfig contains the configuration options for the SyncSecrets operation.
type SyncConfig struct {
	PodNamespace string

	OutOfCluster bool
	KubeConfig   string
}
