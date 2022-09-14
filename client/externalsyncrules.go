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
		return client.ModifiedExternalSyncRuleHandler(rule)
	case watch.Deleted:
		return client.DeletedExternalSyncRuleHandler(rule)
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

// ModifiedExternalSyncRuleHandler handles syncing Services/Ingresses after a ExternalSyncRule has been modified
//
// Due to the event watcher only providing the new state of the modified resource, it is impossible to know the previous state.
// (The exception to this is potentially "applied" changes and parsing the last-applied-configuration annotation)
// In coping with this limitation, a modified ExternalSyncRule will simply attempt to resync the rule across all applicable namespaces.
func (client *Client) ModifiedExternalSyncRuleHandler(rule *typesv1.ExternalSyncRule) error {
	if rule.DeletionTimestamp != nil {
		return nil
	}

	ruleLogger(rule).Infof("modified")

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

func (client *Client) DeletedExternalSyncRuleHandler(rule *typesv1.ExternalSyncRule) error {
	ruleLogger(rule).Infof("deleted")

	var service *v1.Service = nil
	if rule.HasService() && rule.Spec.Service.IsService() {
		service, _ = client.GetService(rule.Spec.Namespace, rule.Spec.Service.Name)
	}

	for _, namespace := range rule.Namespaces(client.Context, client.DefaultClientset) {
		if service != nil {
			client.SyncDeletedExternalNameService(rule, &namespace, service)
		}
	}

	return nil
}
