package ingresses

import (
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

func Logger(service *networkingv1.Ingress, namespaces ...*v1.Namespace) *log.Entry {
	namespace := service.Namespace
	if len(namespaces) > 0 {
		namespace = namespaces[0].Name
	}

	return NameLogger(namespace, service.Name)
}

func NameLogger(namespace, name string) *log.Entry {
	return log.WithFields(log.Fields{"name": name, "kind": "Ingress", "namespace": namespace})
}
