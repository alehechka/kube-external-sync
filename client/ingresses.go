package client

import (
	"reflect"

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

func (client *Client) CreateUpdateIngress(rule *typesv1.ExternalSyncRule, namespace *v1.Namespace, ingress *networkingv1.Ingress) error {
	logger := ingressLogger(ingress, namespace)

	if namespaceIngress, err := client.GetIngress(namespace.Name, ingress.Name); err == nil {
		logger.Debugf("already exists")

		if !IsIngressManagedBy(namespaceIngress) {
			logger.Debugf("existing service is not managed and will not be updated")
			return nil
		}

		if IngressesAreEqual(ingress, namespaceIngress) {
			logger.Debugf("existing service contains same data")
			return nil
		}

		return client.UpdateIngress(rule, namespace, ingress)
	}

	return client.CreateIngress(rule, namespace, ingress)
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

func IngressesAreEqual(a, b *networkingv1.Ingress) bool {
	return (IngressRuleValuesAreEqual(a.Spec.Rules, b.Spec.Rules) &&
		AnnotationsAreEqual(a.Annotations, b.Annotations))
}

func IngressRuleValuesAreEqual(a, b []networkingv1.IngressRule) bool {
	if len(a) != len(b) {
		return false
	}

	for index := range a {
		if !reflect.DeepEqual(a[index].IngressRuleValue, b[index].IngressRuleValue) {
			return false
		}
	}

	return true
}

func PrepareIngress(rule *typesv1.ExternalSyncRule, namespace *v1.Namespace, ingress *networkingv1.Ingress) *networkingv1.Ingress {
	return &networkingv1.Ingress{
		TypeMeta: ingress.TypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:            ingress.Name,
			Namespace:       namespace.Name,
			Labels:          ingress.Labels,
			Annotations:     Manage(CopyAnnotations(ingress.Annotations)),
			OwnerReferences: []metav1.OwnerReference{OwnerReference(rule)},
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: ingress.Spec.IngressClassName,
			DefaultBackend:   ingress.Spec.DefaultBackend,
			TLS:              rule.Spec.Ingress.PrepareTLS(namespace, ingress),
			Rules:            rule.Spec.Ingress.PrepareIngressRules(namespace, ingress),
		},
	}
}

func IsIngressManagedBy(ingress *networkingv1.Ingress) bool {
	managedBy, ok := ingress.Annotations[constants.ManagedByAnnotationKey]

	return ok && managedBy == constants.ManagedByAnnotationValue
}
