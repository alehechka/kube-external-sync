package client

import (
	typesv1 "github.com/alehechka/kube-external-sync/api/types/v1"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

func ruleLogger(rule *typesv1.ExternalSyncRule) *log.Entry {
	return log.WithFields(log.Fields{"name": rule.Name, "kind": "ExternalSyncRule"})
}

func serviceLogger(service *v1.Service) *log.Entry {
	return log.WithFields(log.Fields{"name": service.Name, "kind": "Service", "namespace": service.Namespace})
}

func namespaceLogger(namespace *v1.Namespace) *log.Entry {
	return log.WithFields(log.Fields{"name": namespace.Name, "kind": "Namespace"})
}
