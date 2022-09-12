package client

import (
	log "github.com/sirupsen/logrus"
)

// SyncExternals syncs Services/Ingress across Namespaces as ExternalName references
func SyncExternals(config *SyncConfig) (err error) {
	log.Debugf("Starting with following configuration: %#v", *config)

	client := new(Client)

	if err = client.Initialize(config); err != nil {
		return err
	}

	defer client.ExternalSyncRuleWatcher.Stop()

	for {
		select {
		case externalSyncRuleEvent, ok := <-client.ExternalSyncRuleWatcher.ResultChan():
			if !ok {
				log.Debug("ExternalSyncRule watcher timed out, restarting now.")
				if err := client.StartExternalSyncRuleWatcher(); err != nil {
					return err
				}
				defer client.ExternalSyncRuleWatcher.Stop()
				continue
			}
			client.ExternalSyncRuleEventHandler(externalSyncRuleEvent)
		case s := <-client.SignalChannel:
			log.Infof("Shutting down from signal: %s", s)
			return nil
		}
	}
}
