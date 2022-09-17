package common

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type ReplicatorConfig struct {
	Kind         string
	Client       kubernetes.Interface
	ResyncPeriod time.Duration
	ListFunc     cache.ListFunc
	WatchFunc    cache.WatchFunc
	ObjType      runtime.Object
}

type UpdateFuncs struct {
	ReplicateDataFrom func(source interface{}, target interface{}) error
	ReplicateObjectTo func(source interface{}, target *v1.Namespace) error
}

type GenericReplicator struct {
	ReplicatorConfig
	Store      cache.Store
	Controller cache.Controller

	UpdateFuncs UpdateFuncs

	// ReplicateToList is a set that caches the names of all resources that have a
	// "replicate-to" annotation.
	ReplicateToList map[string]struct{}
}

// NewGenericReplicator creates a new GenericReplicator
func NewGenericReplicator(ctx context.Context, config ReplicatorConfig) *GenericReplicator {
	repl := GenericReplicator{
		ReplicatorConfig: config,
		ReplicateToList:  make(map[string]struct{}),
	}

	store, controller := cache.NewInformer(
		&cache.ListWatch{
			ListFunc:  config.ListFunc,
			WatchFunc: config.WatchFunc,
		},
		config.ObjType,
		config.ResyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    repl.ResourceAdded,
			UpdateFunc: func(old interface{}, new interface{}) { repl.ResourceAdded(new) },
			DeleteFunc: repl.ResourceDeleted,
		},
	)

	namespaceWatcher.OnNamespaceAdded(ctx, config.Client, config.ResyncPeriod, repl.NamespaceAdded)
	namespaceWatcher.OnNamespaceUpdated(ctx, config.Client, config.ResyncPeriod, repl.NamespaceUpdated)

	repl.Store = store
	repl.Controller = controller

	return &repl
}

func (r *GenericReplicator) Synced() bool {
	return r.Controller.HasSynced()
}

func (r *GenericReplicator) Run() {
	log.WithField("kind", r.Kind).Infof("running %s controller", r.Kind)
	r.Controller.Run(wait.NeverStop)
}

// NamespaceAdded replicates resources with ReplicateTo and ReplicateToMatching
// annotations into newly created namespaces.
func (r *GenericReplicator) NamespaceAdded(ns *v1.Namespace) {
	logger := log.WithField("kind", r.Kind).WithField("target", ns.Name)
	logger.Info("NamespaceAdded")
}

// NamespaceUpdated checks if namespace's labels changed and deletes any 'replicate-to-matching' resources
// the namespace no longer qualifies for. Then it attempts to replicate resources into the updated ns based
// on the updated set of labels
func (r *GenericReplicator) NamespaceUpdated(nsOld *v1.Namespace, nsNew *v1.Namespace) {}

// ResourceAdded checks resources with ReplicateTo or ReplicateFromAnnotation annotation
func (r *GenericReplicator) ResourceAdded(obj interface{}) {}

// ResourceDeleted watches for the deletion of resources
func (r *GenericReplicator) ResourceDeleted(source interface{}) {}
