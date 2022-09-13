package client

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (client *Client) GetService(namespace, name string) (service *v1.Service, err error) {
	service, err = client.DefaultClientset.CoreV1().Services(namespace).Get(client.Context, name, metav1.GetOptions{})
	if err != nil {
		serviceLogger(&v1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name}}).
			Errorf("failed to get service: %s", err.Error())
	}
	return
}

func (client *Client) ListServices(namespace string) (list *v1.ServiceList, err error) {
	list, err = client.DefaultClientset.CoreV1().Services(namespace).List(client.Context, metav1.ListOptions{})
	if err != nil {
		namespaceLogger(&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}}).
			Errorf("failed to list services: %s", err.Error())
	}
	return
}
