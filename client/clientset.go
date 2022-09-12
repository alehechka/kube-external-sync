package client

import (
	"github.com/alehechka/kube-external-sync/api/types/v1/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func (client *Client) InitializeClientsets() error {

	if err := client.InitializeClusterConfig(); err != nil {
		return err
	}

	if err := client.InitializeDefault(); err != nil {
		return err
	}

	if err := client.InitializeKubeExternalSync(); err != nil {
		return err
	}

	return nil
}

func (client *Client) InitializeDefault() (err error) {
	client.DefaultClientset, err = kubernetes.NewForConfig(client.ClusterConfig)
	return
}

func (client *Client) InitializeKubeExternalSync() (err error) {
	client.KubeExternalSyncClientset, err = clientset.NewForConfig(client.ClusterConfig)
	return
}

func (client *Client) InitializeClusterConfig() (err error) {
	if client.SyncConfig.OutOfCluster {
		client.ClusterConfig, err = clientcmd.BuildConfigFromFlags("", client.SyncConfig.KubeConfig)
	} else {
		client.ClusterConfig, err = rest.InClusterConfig()
	}

	return
}
