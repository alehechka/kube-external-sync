package common

// ManagedByAnnotationKey is an annotation key appended to kube-external-sync managed resources.
const ManagedByAnnotationKey = "app.kubernetes.io/managed-by"

// ManagedByAnnotationValue is an annotation value appended to kube-external-sync managed resources.
const ManagedByAnnotationValue = "kube-external-sync"

// LastAppliedConfigurationAnnotationKey is an annotation created by Kubernetes to keep track of last config
const LastAppliedConfigurationAnnotationKey = "kubectl.kubernetes.io/last-applied-configuration"
