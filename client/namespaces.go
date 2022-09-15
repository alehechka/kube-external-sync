package client

import (
	typesv1 "github.com/alehechka/kube-external-sync/api/types/v1"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func (client *Client) NamespaceEventHandler(event watch.Event) error {
	namespace, ok := event.Object.(*v1.Namespace)
	if !ok {
		log.Error("failed to cast Namespace")
		return nil
	}

	switch event.Type {
	case watch.Added:
		return client.AddedNamespaceHandler(namespace)
	}

	return nil
}

func (client *Client) AddedNamespaceHandler(namespace *v1.Namespace) error {
	logger := namespaceLogger(namespace)

	if namespace.CreationTimestamp.Time.Before(client.StartTime) {
		logger.Debugf("namespace will be synced on startup by ExternalSyncRule watcher")
		return nil
	}

	logger.Infof("added")
	return client.SyncNamespace(namespace)
}

func (client *Client) SyncNamespace(namespace *v1.Namespace) error {
	namespaceLogger(namespace).Debugf("syncing new namespace")

	rules, err := client.ListExternalSyncRules()
	if err != nil {
		return err
	}

	for _, rule := range rules.Items {
		if rule.ShouldSyncNamespace(namespace) {
			client.SyncResourcesToNamespace(namespace, &rule)
		}
	}

	return nil
}

func (client *Client) SyncResourcesToNamespace(namespace *v1.Namespace, rule *typesv1.ExternalSyncRule) error {
	if rule.HasService() && rule.Spec.Service.IsService() {
		if service, err := client.GetService(rule.Spec.Namespace, rule.Spec.Service.Name); service != nil && err == nil {
			client.CreateUpdateExternalNameService(rule, namespace, service)
		}
	}

	return nil
}

func (client *Client) ListNamespaces() (namespaces *v1.NamespaceList, err error) {
	namespaces, err = client.DefaultClientset.CoreV1().Namespaces().List(client.Context, metav1.ListOptions{})
	if err != nil {
		log.Errorf("failed to list namespaces: %s", err.Error())
	}
	return
}
