package client

import (
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (client *Client) ListNamespaces() (namespaces *v1.NamespaceList, err error) {
	namespaces, err = client.DefaultClientset.CoreV1().Namespaces().List(client.Context, metav1.ListOptions{})
	if err != nil {
		log.Errorf("failed to list namespaces: %s", err.Error())
	}
	return
}
