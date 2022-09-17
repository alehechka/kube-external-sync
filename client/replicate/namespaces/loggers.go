package namespaces

import (
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

func namespaceLogger(namespace *v1.Namespace) *log.Entry {
	return namespaceNameLogger(namespace.Name)
}

func namespaceNameLogger(name string) *log.Entry {
	return log.WithFields(log.Fields{"name": name, "kind": "Namespace"})
}
