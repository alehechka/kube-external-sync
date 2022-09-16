package client

import (
	"github.com/alehechka/kube-external-sync/constants"
	log "github.com/sirupsen/logrus"
	networkingv1 "k8s.io/api/networking/v1"
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

func IsIngressManagedBy(ingress *networkingv1.Ingress) bool {
	managedBy, ok := ingress.Annotations[constants.ManagedByAnnotationKey]

	return ok && managedBy == constants.ManagedByAnnotationValue
}
