package client

// SyncConfig contains the configuration options for the SyncSecrets operation.
type SyncConfig struct {
	PodNamespace string

	LivenessPort int

	OutOfCluster bool
	KubeConfig   string
}
