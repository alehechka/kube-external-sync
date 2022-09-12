package client

import (
	typesv1 "github.com/alehechka/kube-external-sync/api/types/v1"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/watch"
)

func (client *Client) ExternalSyncRuleEventHandler(event watch.Event) error {
	rule, ok := event.Object.(*typesv1.ExternalSyncRule)
	if !ok {
		log.Error("failed to cast ExternalSyncRule")
		return nil
	}

	switch event.Type {
	case watch.Added:
		ruleLogger(rule).Infof("added, HasService: %v, HasIngress: %v, IncludeAnnotations: %#v", rule.HasService(), rule.HasIngress(), rule.Spec.Rules.Namespaces.IncludeAnnotation)
	case watch.Modified:
		ruleLogger(rule).Infof("modified: %#v", rule)
	case watch.Deleted:
		ruleLogger(rule).Infof("deleted: %#v", rule)
	}

	return nil
}
