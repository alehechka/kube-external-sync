package client

import (
	typesv1 "github.com/alehechka/kube-external-sync/api/types/v1"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

func ruleLogger(rule *typesv1.ExternalSyncRule) *log.Entry {
	return log.WithFields(log.Fields{"name": rule.Name, "kind": "ExternalSyncRule"})
}

func secretLogger(secret *v1.Secret) *log.Entry {
	return log.WithFields(log.Fields{"name": secret.Name, "kind": "Secret", "namespace": secret.Namespace})
}

func namespaceLogger(namespace *v1.Namespace) *log.Entry {
	return log.WithFields(log.Fields{"name": namespace.Name, "kind": "Namespace"})
}
