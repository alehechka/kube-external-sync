package client

import (
	"reflect"

	typesv1 "github.com/alehechka/kube-external-sync/api/types/v1"
	"github.com/alehechka/kube-external-sync/api/types/v1/clientset"
	"github.com/alehechka/kube-external-sync/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
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

func OwnerReference(rule *typesv1.ExternalSyncRule) metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion: clientset.GroupName,
		Kind:       clientset.ExternalSyncRule,
		Name:       rule.Name,
		UID:        uuid.NewUUID(),
	}
}
