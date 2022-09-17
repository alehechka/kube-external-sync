package client

import (
	"github.com/alehechka/kube-external-sync/client/liveness"
	"github.com/alehechka/kube-external-sync/client/replicate/common"
	log "github.com/sirupsen/logrus"
)

// SyncExternals syncs Services/Ingress across Namespaces as ExternalName references
func SyncExternals(config *SyncConfig) (err error) {
	log.Debugf("Starting with following configuration: %#v", *config)

	return liveness.Serve(config.LivenessPort, []common.Replicator{})
}
