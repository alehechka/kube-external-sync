package client

import (
	"context"
	"time"

	"github.com/alehechka/kube-external-sync/client/replicate/common"
	"github.com/alehechka/kube-external-sync/client/replicate/ingress"
	"github.com/alehechka/kube-external-sync/client/replicate/service"
	"github.com/alehechka/kube-external-sync/client/replicate/traefik/ingressroute"
	"github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/generated/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// SyncConfig contains the configuration options for the SyncExternals operation.
type SyncConfig struct {
	PodNamespace string

	LivenessPort           int
	ResyncPeriod           time.Duration
	DefaultIngressHostname string
	EnableTraefik          bool

	OutOfCluster bool
	KubeConfig   string
}

type Controller struct {
	SyncConfig *SyncConfig
	Context    context.Context

	ClientConfig  *rest.Config
	DefaultClient kubernetes.Interface
	TraefikClient *versioned.Clientset

	ServiceReplicator             common.Replicator
	IngressReplicator             common.Replicator
	TraefikIngressRouteReplicator common.Replicator
}

func NewController() *Controller {
	return new(Controller)
}

func (c *Controller) Initialize(config *SyncConfig) (*Controller, error) {
	controller := new(Controller)

	controller.SyncConfig = config
	controller.Context = context.Background()

	if err := controller.InitializeClients(); err != nil {
		return nil, err
	}

	controller.InitializeReplicators()

	return controller, nil
}

func (c *Controller) InitializeClients() (err error) {
	if err := c.InitializeClusterConfig(); err != nil {
		return err
	}

	if err := c.InitializeDefaultClient(); err != nil {
		return err
	}

	if c.SyncConfig.EnableTraefik {
		if err := c.InitializeTraefikClient(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) InitializeDefaultClient() (err error) {
	c.DefaultClient, err = kubernetes.NewForConfig(c.ClientConfig)
	return
}

func (c *Controller) InitializeTraefikClient() (err error) {
	c.TraefikClient, err = versioned.NewForConfig(c.ClientConfig)
	return err
}

func (c *Controller) InitializeClusterConfig() (err error) {
	if c.SyncConfig.OutOfCluster {
		c.ClientConfig, err = clientcmd.BuildConfigFromFlags("", c.SyncConfig.KubeConfig)
		return
	}

	c.ClientConfig, err = rest.InClusterConfig()
	return
}

func (c *Controller) InitializeReplicators() {
	c.ServiceReplicator = service.NewReplicator(c.Context, c.DefaultClient, c.SyncConfig.ResyncPeriod)
	c.IngressReplicator = ingress.NewReplicator(c.Context, c.DefaultClient, c.SyncConfig.ResyncPeriod, c.SyncConfig.DefaultIngressHostname)

	if c.SyncConfig.EnableTraefik {
		c.TraefikIngressRouteReplicator = ingressroute.NewReplicator(c.Context, c.DefaultClient, c.TraefikClient, c.SyncConfig.ResyncPeriod, c.SyncConfig.DefaultIngressHostname)
	}
}
