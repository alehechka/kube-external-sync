package client

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (client *Client) InitializeWatchers() (err error) {
	if err := client.StartSecretWatcher(); err != nil {
		return err
	}

	if err := client.StartNamespaceWatcher(); err != nil {
		return err
	}

	if err := client.StartExternalSyncRuleWatcher(); err != nil {
		return err
	}

	return
}

func (client *Client) StartSecretWatcher() (err error) {
	client.SecretWatcher, err = client.DefaultClientset.CoreV1().Secrets(v1.NamespaceAll).Watch(client.Context, metav1.ListOptions{})
	return
}

func (client *Client) StartNamespaceWatcher() (err error) {
	client.NamespaceWatcher, err = client.DefaultClientset.CoreV1().Namespaces().Watch(client.Context, metav1.ListOptions{})
	return
}

func (client *Client) StartExternalSyncRuleWatcher() (err error) {
	client.ExternalSyncRuleWatcher, err = client.KubeExternalSyncClientset.ExternalSyncRules().Watch(client.Context, metav1.ListOptions{})
	return
}
