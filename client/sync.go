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
	defer client.NamespaceWatcher.Stop()
	defer client.ServiceWatcher.Stop()

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
		case namespaceEvent, ok := <-client.NamespaceWatcher.ResultChan():
			if !ok {
				log.Debug("Namespace watcher timed out, restarting now.")
				if err := client.StartNamespaceWatcher(); err != nil {
					return err
				}
				defer client.NamespaceWatcher.Stop()
				continue
			}
			client.NamespaceEventHandler(namespaceEvent)
		case serviceEvent, ok := <-client.ServiceWatcher.ResultChan():
			if !ok {
				log.Debug("Service watcher timed out, restarting now.")
				if err := client.StartServiceWatcher(); err != nil {
					return err
				}
				defer client.ServiceWatcher.Stop()
				continue
			}
			client.ServiceEventHandler(serviceEvent)
		case s := <-client.SignalChannel:
			log.Infof("Shutting down from signal: %s", s)
			return nil
		}
	}
}
