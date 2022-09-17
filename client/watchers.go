package client

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (client *Client) InitializeWatchers() (err error) {

	if err := client.StartExternalSyncRuleWatcher(); err != nil {
		return err
	}

	if err := client.StartNamespaceWatcher(); err != nil {
		return err
	}

	if err := client.StartServiceWatcher(); err != nil {
		return err
	}

	if err := client.StartIngressWatcher(); err != nil {
		return err
	}

	return
}

func (client *Client) StartExternalSyncRuleWatcher() (err error) {
	client.ExternalSyncRuleWatcher, err = client.KubeExternalSyncClientset.ExternalSyncRules().Watch(client.Context, metav1.ListOptions{})
	return
}

func (client *Client) StartNamespaceWatcher() (err error) {
	client.NamespaceWatcher, err = client.DefaultClientset.CoreV1().Namespaces().Watch(client.Context, metav1.ListOptions{})
	return
}

func (client *Client) StartServiceWatcher() (err error) {
	client.ServiceWatcher, err = client.DefaultClientset.CoreV1().Services(v1.NamespaceAll).Watch(client.Context, metav1.ListOptions{})
	return
}

func (client *Client) StartIngressWatcher() (err error) {
	client.IngressWatcher, err = client.DefaultClientset.NetworkingV1().Ingresses(v1.NamespaceAll).Watch(client.Context, metav1.ListOptions{})
	return
}
