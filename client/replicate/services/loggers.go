package services

import (
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

func Logger(service *v1.Service, namespaces ...*v1.Namespace) *log.Entry {
	namespace := service.Namespace
	if len(namespaces) > 0 {
		namespace = namespaces[0].Name
	}

	return NameLogger(namespace, service.Name)
}

func NameLogger(namespace, name string) *log.Entry {
	return log.WithFields(log.Fields{"name": name, "kind": "Service", "namespace": namespace})
}
