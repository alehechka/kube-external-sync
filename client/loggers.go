package client

import (
	typesv1 "github.com/alehechka/kube-external-sync/api/types/v1"
	log "github.com/sirupsen/logrus"
)

func ruleLogger(rule *typesv1.ExternalSyncRule) *log.Entry {
	return log.WithFields(log.Fields{"name": rule.Name, "kind": "ExternalSyncRule"})
}
