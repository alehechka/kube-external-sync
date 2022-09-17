package client

import (
	"reflect"

	typesv1 "github.com/alehechka/kube-external-sync/api/types/v1"
	"github.com/alehechka/kube-external-sync/constants"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func (client *Client) ServiceEventHandler(event watch.Event) error {
	service, ok := event.Object.(*v1.Service)
	if !ok {
		log.Error("failed to cast Service")
		return nil
	}

	if IsServiceManagedBy(service) {
		return nil
	}

	switch event.Type {
	case watch.Added:
		return client.AddedServiceHandler(service)
	case watch.Modified:
		return client.ModifiedServiceHandler(service)
	case watch.Deleted:
		return client.DeletedServiceHandler(service)
	}

	return nil
}

func (client *Client) AddedServiceHandler(service *v1.Service) error {
	logger := serviceLogger(service)

	if service.CreationTimestamp.Time.Before(client.StartTime) {
		logger.Debugf("service will be synced on startup by ExternalSyncRule watcher")
		return nil
	}

	logger.Infof("added")
	return client.SyncAddedModifiedService(service)
}

func (client *Client) ModifiedServiceHandler(service *v1.Service) error {
	if service.DeletionTimestamp != nil {
		return nil
	}

	serviceLogger(service).Infof("modified")
	return client.SyncAddedModifiedService(service)
}

func (client *Client) SyncAddedModifiedService(service *v1.Service) error {
	rules, err := client.ListExternalSyncRules()
	if err != nil {
		return err
	}

	for _, rule := range rules.Items {
		if rule.ShouldSyncService(service) {
			for _, namespace := range rule.Namespaces(client.Context, client.DefaultClientset) {
				client.CreateUpdateExternalNameService(&rule, &namespace, service)
			}
		}
	}

	return nil
}

func (client *Client) CreateUpdateExternalNameService(rule *typesv1.ExternalSyncRule, namespace *v1.Namespace, service *v1.Service) error {
	logger := serviceLogger(service, namespace)

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

func (client *Client) DeletedServiceHandler(service *v1.Service) error {
	serviceLogger(service).Infof("deleted")

	rules, err := client.ListExternalSyncRules()
	if err != nil {
		return err
	}

	for _, rule := range rules.Items {
		if rule.ShouldSyncService(service) {
			for _, namespace := range rule.Namespaces(client.Context, client.DefaultClientset) {
				client.SyncDeletedService(&namespace, service)
			}
		}
	}

	return nil
}

func (client *Client) SyncDeletedService(namespace *v1.Namespace, service *v1.Service) error {
	logger := serviceLogger(service, namespace)

	if namespaceService, err := client.GetService(namespace.Name, service.Name); err == nil {
		if IsServiceManagedBy(namespaceService) {
			return client.DeleteService(namespace, service)
		}

		logger.Debugf("existing service is not managed and will not be deleted")
	}

	return nil
}

func (client *Client) GetService(namespace, name string) (service *v1.Service, err error) {
	service, err = client.DefaultClientset.CoreV1().Services(namespace).Get(client.Context, name, metav1.GetOptions{})
	if err != nil {
		serviceNameLogger(namespace, name).
			Errorf("failed to get service: %s", err.Error())
	}
	return
}

func (client *Client) ListServices(namespace string) (list *v1.ServiceList, err error) {
	list, err = client.DefaultClientset.CoreV1().Services(namespace).List(client.Context, metav1.ListOptions{})
	if err != nil {
		namespaceNameLogger(namespace).
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

func (client *Client) DeleteService(namespace *v1.Namespace, service *v1.Service) (err error) {
	logger := serviceLogger(service, namespace)

	logger.Infof("deleting service")

	err = client.DefaultClientset.CoreV1().Services(namespace.Name).Delete(client.Context, service.Name, metav1.DeleteOptions{})
	if err != nil {
		logger.Errorf("failed to delete service - %s", err.Error())
	}

	return
}

func ExternalNameServicesAreEqual(a, b *v1.Service) bool {
	return (reflect.DeepEqual(a.Spec.Ports, b.Spec.Ports) &&
		AnnotationsAreEqual(a.Annotations, b.Annotations))
}

func PrepareExternalNameService(rule *typesv1.ExternalSyncRule, namespace *v1.Namespace, service *v1.Service) *v1.Service {
	annotations := Manage(CopyAnnotations(service.Annotations))

	return &v1.Service{
		TypeMeta: service.TypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:            service.Name,
			Namespace:       namespace.Name,
			Labels:          service.Labels,
			Annotations:     annotations,
			OwnerReferences: []metav1.OwnerReference{OwnerReference(rule)},
		},
		Spec: v1.ServiceSpec{
			Type:         v1.ServiceTypeExternalName,
			ExternalName: rule.ServiceExternalName(),
			Ports:        service.Spec.Ports,
		},
	}
}

func IsServiceManagedBy(service *v1.Service) bool {
	managedBy, ok := service.Annotations[constants.ManagedByAnnotationKey]

	return ok && managedBy == constants.ManagedByAnnotationValue
}
