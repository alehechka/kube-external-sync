package common

import v1 "k8s.io/api/core/v1"

type Replicator interface {
	Run()
	Synced() bool
	NamespaceAdded(ns *v1.Namespace)
}
