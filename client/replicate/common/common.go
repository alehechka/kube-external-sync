package common

import (
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Replicator interface {
	Run()
	Synced() bool
	NamespaceAdded(ns *v1.Namespace)
}

// CopyAnnotations copies all non-controlled annotations
func CopyAnnotations(m map[string]string) map[string]string {
	copy := make(map[string]string)

	for key, value := range m {
		if _, ok := DefaultStripAnnotations[key]; ok {
			continue
		}
		copy[key] = value
	}

	return copy
}

// PrepareAnnotations prepares a new map of annotations based on the provided resource
func PrepareAnnotations(source metav1.ObjectMeta) map[string]string {
	annotations := make(map[string]string)
	if stripAnnotations, ok := source.Annotations[StripAnnotations]; !ok || stripAnnotations != "true" {
		annotations = CopyAnnotations(source.Annotations)
	}

	annotations[ReplicatedFromAnnotation] = fmt.Sprintf("%s/%s", source.Namespace, source.Name)
	annotations[ReplicatedAtAnnotation] = time.Now().Format(time.RFC3339)
	annotations[ReplicatedFromVersionAnnotation] = source.ResourceVersion

	return annotations
}

// CopyLabels copies all non-controlled Labels
func CopyLabels(m map[string]string) map[string]string {
	copy := make(map[string]string)

	for key, value := range m {
		if key == ManagedByLabelKey {
			continue
		}
		copy[key] = value
	}

	return copy
}

// PrepareLabels prepares a new map of labels based on the provided resource
func PrepareLabels(source metav1.ObjectMeta) map[string]string {
	labels := make(map[string]string)
	if stripLabels, ok := source.Labels[StripLabels]; !ok || stripLabels != "true" {
		labels = CopyLabels(source.Labels)
	}

	labels[ManagedByLabelKey] = ManagedByLabelValue

	return labels
}

// IsManagedBy checks the sources labels to see if the resource is currently being managed-by this controller
func IsManagedBy(source metav1.Object) bool {
	managedBy, ok := source.GetLabels()[ManagedByLabelKey]

	return ok && managedBy == ManagedByLabelValue
}

// PrepareOwnerReferences prepares the OwnerReferences array
func PrepareOwnerReferences(source metav1.ObjectMeta) []metav1.OwnerReference {
	keepOwnerReferences, ok := source.Annotations[KeepOwnerReferences]
	if ok && keepOwnerReferences == "true" {
		return source.OwnerReferences
	}

	return nil
}
