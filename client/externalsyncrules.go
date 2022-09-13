package client

import (
	typesv1 "github.com/alehechka/kube-external-sync/api/types/v1"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
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
		return client.AddedExternalSyncRuleHandler(rule)
	case watch.Modified:
		ruleLogger(rule).Infof("modified: %#v", rule)
	case watch.Deleted:
		ruleLogger(rule).Infof("deleted: %#v", rule)
	}

	return nil
}

func (client *Client) AddedExternalSyncRuleHandler(rule *typesv1.ExternalSyncRule) (err error) {
	ruleLogger(rule).Infof("added")

	var service *v1.Service = nil
	if rule.HasService() && rule.Spec.Service.IsService() {
		service, _ = client.GetService(rule.Spec.Namespace, rule.Spec.Service.Name)
	}

	for _, namespace := range rule.Namespaces(client.Context, client.DefaultClientset) {
		if service != nil {
			client.CreateUpdateExternalNameService(rule, &namespace, service)
		}
	}

	return nil
}
