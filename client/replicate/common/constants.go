package common

// ManagedByLabelKey is an label key appended to kube-external-sync managed resources.
const ManagedByLabelKey = "app.kubernetes.io/managed-by"

// ManagedByLabelValue is an label value appended to kube-external-sync managed resources.
const ManagedByLabelValue = "kube-external-sync"

// LastAppliedConfigurationAnnotationKey is an annotation created by Kubernetes to keep track of last config
const LastAppliedConfigurationAnnotationKey = "kubectl.kubernetes.io/last-applied-configuration"

// Annotations that are added to resources and used by this Controller
const (
	ReplicateTo         = "kube-external-sync.io/replicate-to"
	ReplicateToMatching = "kube-external-sync.io/replicate-to-matching"
	StripLabels         = "kube-external-sync.io/strip-labels"
	StripAnnotations    = "kube-external-sync.io/strip-annotations"
	TopLevelDomain      = "kube-external-sync.io/top-level-domain"
	TLDSecretName       = "kube-external-sync.io/tld-secret-name"
	ExternalNameSuffix  = "kube-external-sync.io/external-name-suffix"
	KeepOwnerReferences = "kube-external-sync.io/keep-owner-references"
)

// Annotations that are added to replicated resources by this Controller
const (
	ReplicatedFromAnnotation        = "kube-external-sync.io/replicated-from"
	ReplicatedAtAnnotation          = "kube-external-sync.io/replicated-at"
	ReplicatedFromVersionAnnotation = "kube-external-sync.io/replicated-from-version"
)

// DefaultStripAnnotations contains the annotations that are to be stripped when replicating a resource
var DefaultStripAnnotations = map[string]struct{}{
	LastAppliedConfigurationAnnotationKey: {},
	ReplicateTo:                           {},
	ReplicateToMatching:                   {},
	StripLabels:                           {},
	StripAnnotations:                      {},
	TopLevelDomain:                        {},
	TLDSecretName:                         {},
}

// ExternalName suffix options
const (
	DefaultExternalNameSuffix     = "svc.cluster.local"
	TraefikMeshExternalNameSuffix = "traefik.mesh"
)
