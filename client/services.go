package client

import (
	typesv1 "github.com/alehechka/kube-external-sync/api/types/v1"
	"github.com/alehechka/kube-external-sync/constants"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (client *Client) CreateUpdateExternalNameService(rule *typesv1.ExternalSyncRule, namespace *v1.Namespace, service *v1.Service) error {
	externalService := PrepareExternalNameService(rule, namespace, service)
	serviceLogger(externalService).Infof("creating: %#v", externalService)
	return nil
}

func (client *Client) GetService(namespace, name string) (service *v1.Service, err error) {
	service, err = client.DefaultClientset.CoreV1().Services(namespace).Get(client.Context, name, metav1.GetOptions{})
	if err != nil {
		serviceLogger(&v1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name}}).
			Errorf("failed to get service: %s", err.Error())
	}
	return
}

func (client *Client) ListServices(namespace string) (list *v1.ServiceList, err error) {
	list, err = client.DefaultClientset.CoreV1().Services(namespace).List(client.Context, metav1.ListOptions{})
	if err != nil {
		namespaceLogger(&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}}).
			Errorf("failed to list services: %s", err.Error())
	}
	return
}

func PrepareExternalNameService(rule *typesv1.ExternalSyncRule, namespace *v1.Namespace, service *v1.Service) *v1.Service {
	annotations := CopyAnnotations(service.Annotations)
	annotations[constants.ManagedByAnnotationKey] = constants.ManagedByAnnotationValue

	return &v1.Service{
		TypeMeta: service.TypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:        service.Name,
			Namespace:   namespace.Name,
			Labels:      service.Labels,
			Annotations: annotations,
		},
		Spec: v1.ServiceSpec{
			Type:         v1.ServiceTypeExternalName,
			ExternalName: rule.ServiceExternalName(),
			Ports:        service.Spec.Ports,
		},
	}
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

func IsManagedBy(service *v1.Service) bool {
	managedBy, ok := service.Annotations[constants.ManagedByAnnotationKey]

	return ok && managedBy == constants.ManagedByAnnotationValue
}
