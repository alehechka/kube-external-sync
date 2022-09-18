package client

import (
	"context"

	"github.com/alehechka/kube-external-sync/client/liveness"
	"github.com/alehechka/kube-external-sync/client/replicate/common"
	"github.com/alehechka/kube-external-sync/client/replicate/ingress"
	"github.com/alehechka/kube-external-sync/client/replicate/service"
	log "github.com/sirupsen/logrus"
)

// SyncExternals syncs Services/Ingress across Namespaces as ExternalName references
func SyncExternals(config *SyncConfig) (err error) {
	log.Debugf("Starting with following configuration: %#v", *config)

	client, err := InitializeClient(config)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	serviceRepl := service.NewReplicator(ctx, client, config.ResyncPeriod)
	ingressRepl := ingress.NewReplicator(ctx, client, config.ResyncPeriod)

	go serviceRepl.Run()
	go ingressRepl.Run()

	return liveness.Serve(config.LivenessPort, []common.Replicator{serviceRepl, ingressRepl})
}
