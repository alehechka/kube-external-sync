package client

import (
	"time"

	log "github.com/sirupsen/logrus"
)

// SyncExternals syncs Services/Ingress across Namespaces as ExternalName references
func SyncExternals(config *SyncConfig) (err error) {
	for {
		log.Info("running...")
		time.Sleep(time.Second * 10)
	}
}
