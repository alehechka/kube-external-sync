package client

import (
	"github.com/alehechka/kube-external-sync/client/liveness"
	"github.com/alehechka/kube-external-sync/client/replicate/common"
	log "github.com/sirupsen/logrus"
)

// SyncExternals syncs Services/Ingress across Namespaces as ExternalName references
func SyncExternals(config *SyncConfig) (err error) {
	log.Debugf("Starting with following configuration: %#v", *config)

	controller, err := NewController().Initialize(config)
	if err != nil {
		return err
	}

	go controller.ServiceReplicator.Run()
	go controller.IngressReplicator.Run()

	if config.EnableTraefik {
		go controller.TraefikIngressRouteReplicator.Run()
	}

	return liveness.Serve(config.LivenessPort, []common.Replicator{controller.ServiceReplicator, controller.IngressReplicator, controller.TraefikIngressRouteReplicator})
}
