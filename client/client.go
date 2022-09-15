package client

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	kesclientset "github.com/alehechka/kube-external-sync/api/types/v1/clientset"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Client struct {
	SyncConfig    *SyncConfig
	ClusterConfig *rest.Config
	Context       context.Context
	StartTime     time.Time

	DefaultClientset          kubernetes.Interface
	KubeExternalSyncClientset *kesclientset.KubeExternalSyncClientset

	ExternalSyncRuleWatcher watch.Interface
	NamespaceWatcher        watch.Interface
	SignalChannel           chan os.Signal
}

func (client *Client) Initialize(config *SyncConfig) error {
	client.SyncConfig = config
	client.Context = context.Background()
	client.StartTime = time.Now()

	if err := client.InitializeClientsets(); err != nil {
		return err
	}

	if err := client.InitializeWatchers(); err != nil {
		return err
	}

	client.InitializeSignalChannel()

	return nil
}

func (client *Client) InitializeSignalChannel() {
	client.SignalChannel = make(chan os.Signal, 1)
	signal.Notify(client.SignalChannel, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
}
