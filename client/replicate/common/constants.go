package common

// ManagedByLabelKey is an label key appended to kube-external-sync managed resources.
const ManagedByLabelKey = "app.kubernetes.io/managed-by"

// ManagedByLabelValue is an label value appended to kube-external-sync managed resources.
const ManagedByLabelValue = "kube-external-sync"

// LastAppliedConfigurationAnnotationKey is an annotation created by Kubernetes to keep track of last config
const LastAppliedConfigurationAnnotationKey = "kubectl.kubernetes.io/last-applied-configuration"

// Annotations that are used by this Controller
const (
	ReplicateTo                     = "kube-external-sync.io/replicate-to"
	ReplicateToMatching             = "kube-external-sync.io/replicate-to-matching"
	ReplicatedFromAnnotation        = "kube-external-sync.io/replicated-from"
	ReplicatedAtAnnotation          = "kube-external-sync.io/replicated-at"
	ReplicatedFromVersionAnnotation = "kube-external-sync.io/replicated-from-version"
)
