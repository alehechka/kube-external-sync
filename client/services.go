package client

import (
	"reflect"

	typesv1 "github.com/alehechka/kube-external-sync/api/types/v1"
	"github.com/alehechka/kube-external-sync/constants"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (client *Client) CreateUpdateExternalNameService(rule *typesv1.ExternalSyncRule, namespace *v1.Namespace, service *v1.Service) error {
	logger := serviceLogger(PrepareExternalNameService(rule, namespace, service))

	if namespaceService, err := client.GetService(namespace.Name, service.Name); err == nil {
		logger.Debugf("already exists")

		if !IsServiceManagedBy(namespaceService) {
			logger.Debugf("existing service is not managed and will not be updated")
			return nil
		}

		if ExternalNameServicesAreEqual(service, namespaceService) {
			logger.Debugf("existing service contains same data")
			return nil
		}

		return client.UpdateExternalNameService(rule, namespace, service)
	}

	return client.CreateExternalNameService(rule, namespace, service)
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

func (client *Client) CreateExternalNameService(rule *typesv1.ExternalSyncRule, namespace *v1.Namespace, service *v1.Service) error {
	newService := PrepareExternalNameService(rule, namespace, service)

	logger := serviceLogger(newService)
	logger.Infof("creating ExternalName service")

	_, err := client.DefaultClientset.CoreV1().Services(namespace.Name).Create(client.Context, newService, metav1.CreateOptions{})

	if err != nil {
		logger.Errorf("failed to create service - %s", err.Error())
	}

	return err
}

func (client *Client) UpdateExternalNameService(rule *typesv1.ExternalSyncRule, namespace *v1.Namespace, service *v1.Service) error {
	updateService := PrepareExternalNameService(rule, namespace, service)

	logger := serviceLogger(updateService)
	logger.Infof("updating ExternalName service")

	_, err := client.DefaultClientset.CoreV1().Services(namespace.Name).Update(client.Context, updateService, metav1.UpdateOptions{})

	if err != nil {
		logger.Errorf("failed to update service - %s", err.Error())
	}

	return err
}

func ExternalNameServicesAreEqual(a, b *v1.Service) bool {
	return (reflect.DeepEqual(a.Spec.Ports, b.Spec.Ports) &&
		AnnotationsAreEqual(a.Annotations, b.Annotations))
}

func AnnotationsAreEqual(a, b map[string]string) bool {
	aCopy := CopyAnnotations(a)
	bCopy := CopyAnnotations(b)

	return reflect.DeepEqual(aCopy, bCopy)
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

func IsServiceManagedBy(service *v1.Service) bool {
	managedBy, ok := service.Annotations[constants.ManagedByAnnotationKey]

	return ok && managedBy == constants.ManagedByAnnotationValue
}
