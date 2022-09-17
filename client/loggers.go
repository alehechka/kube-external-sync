package client

import (
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

func serviceLogger(service *v1.Service, namespaces ...*v1.Namespace) *log.Entry {
	namespace := service.Namespace
	if len(namespaces) > 0 {
		namespace = namespaces[0].Name
	}

	return serviceNameLogger(namespace, service.Name)
}

func serviceNameLogger(namespace, name string) *log.Entry {
	return log.WithFields(log.Fields{"name": name, "kind": "Service", "namespace": namespace})
}

func namespaceLogger(namespace *v1.Namespace) *log.Entry {
	return namespaceNameLogger(namespace.Name)
}

func namespaceNameLogger(name string) *log.Entry {
	return log.WithFields(log.Fields{"name": name, "kind": "Namespace"})
}

func ingressLogger(service *networkingv1.Ingress, namespaces ...*v1.Namespace) *log.Entry {
	namespace := service.Namespace
	if len(namespaces) > 0 {
		namespace = namespaces[0].Name
	}

	return ingressNameLogger(namespace, service.Name)
}

func ingressNameLogger(namespace, name string) *log.Entry {
	return log.WithFields(log.Fields{"name": name, "kind": "Ingress", "namespace": namespace})
}
