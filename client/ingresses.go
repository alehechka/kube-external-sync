package client

import (
	typesv1 "github.com/alehechka/kube-external-sync/api/types/v1"
	"github.com/alehechka/kube-external-sync/constants"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func (client *Client) IngressEventHandler(event watch.Event) error {
	ingress, ok := event.Object.(*networkingv1.Ingress)
	if !ok {
		log.Error("failed to cast Ingress")
		return nil
	}

	if IsIngressManagedBy(ingress) {
		return nil
	}

	switch event.Type {
	case watch.Added:
		ingressLogger(ingress).Infof("added")
		// return client.AddedIngressHandler(ingress)
	case watch.Modified:
		ingressLogger(ingress).Infof("modified")
		// return client.ModifiedIngressHandler(ingress)
	case watch.Deleted:
		ingressLogger(ingress).Infof("deleted")
		// return client.DeletedIngressHandler(ingress)
	}

	return nil
}

func (client *Client) GetIngress(namespace, name string) (ingress *networkingv1.Ingress, err error) {
	ingress, err = client.DefaultClientset.NetworkingV1().Ingresses(namespace).Get(client.Context, name, metav1.GetOptions{})
	if err != nil {
		ingressNameLogger(namespace, name).
			Errorf("failed to get ingress: %s", err.Error())
	}
	return
}

func (client *Client) ListIngresses(namespace string) (list *networkingv1.IngressList, err error) {
	list, err = client.DefaultClientset.NetworkingV1().Ingresses(namespace).List(client.Context, metav1.ListOptions{})
	if err != nil {
		namespaceNameLogger(namespace).
			Errorf("failed to list ingresses: %s", err.Error())
	}
	return
}

func (client *Client) CreateIngress(rule *typesv1.ExternalSyncRule, namespace *v1.Namespace, ingress *networkingv1.Ingress) error {
	newIngress := PrepareIngress(rule, namespace, ingress)

	logger := ingressLogger(newIngress)
	logger.Infof("creating")

	_, err := client.DefaultClientset.NetworkingV1().Ingresses(namespace.Name).Create(client.Context, newIngress, metav1.CreateOptions{})

	if err != nil {
		logger.Errorf("failed to create ingress - %s", err.Error())
	}

	return err
}

func (client *Client) UpdateIngress(rule *typesv1.ExternalSyncRule, namespace *v1.Namespace, ingress *networkingv1.Ingress) error {
	updateIngress := PrepareIngress(rule, namespace, ingress)

	logger := ingressLogger(updateIngress)
	logger.Infof("updating")

	_, err := client.DefaultClientset.NetworkingV1().Ingresses(namespace.Name).Update(client.Context, updateIngress, metav1.UpdateOptions{})

	if err != nil {
		logger.Errorf("failed to update service - %s", err.Error())
	}

	return err
}

func (client *Client) DeleteIngress(namespace *v1.Namespace, ingress *networkingv1.Ingress) (err error) {
	logger := ingressLogger(ingress, namespace)

	logger.Infof("deleting")

	err = client.DefaultClientset.NetworkingV1().Ingresses(namespace.Name).Delete(client.Context, ingress.Name, metav1.DeleteOptions{})
	if err != nil {
		logger.Errorf("failed to delete ingress - %s", err.Error())
	}

	return
}

func PrepareIngress(rule *typesv1.ExternalSyncRule, namespace *v1.Namespace, service *networkingv1.Ingress) *networkingv1.Ingress {
	annotations := Manage(CopyAnnotations(service.Annotations))

	return &networkingv1.Ingress{
		TypeMeta: service.TypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:        service.Name,
			Namespace:   namespace.Name,
			Labels:      service.Labels,
			Annotations: annotations,
		},
		Spec: networkingv1.IngressSpec{},
	}
}

func IsIngressManagedBy(ingress *networkingv1.Ingress) bool {
	managedBy, ok := ingress.Annotations[constants.ManagedByAnnotationKey]

	return ok && managedBy == constants.ManagedByAnnotationValue
}
