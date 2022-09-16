package client

import (
	"reflect"

	"github.com/alehechka/kube-external-sync/constants"
)

func AnnotationsAreEqual(a, b map[string]string) bool {
	aCopy := CopyAnnotations(a)
	bCopy := CopyAnnotations(b)

	return reflect.DeepEqual(aCopy, bCopy)
}

func CopyAnnotations(m map[string]string) map[string]string {
	copy := make(map[string]string)

	for key, value := range m {
		if key == constants.ManagedByAnnotationKey || key == constants.LastAppliedConfigurationAnnotationKey {
			continue
		}
		copy[key] = value
	}

	return copy
}

func Manage(m map[string]string) map[string]string {
	if m == nil {
		m = make(map[string]string)
	}
	m[constants.ManagedByAnnotationKey] = constants.ManagedByAnnotationValue
	return m
}
