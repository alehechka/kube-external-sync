package common

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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
	ReplicateDataFrom        func(source interface{}, target interface{}) error
	ReplicateObjectTo        func(source interface{}, target *v1.Namespace) error
	DeleteReplicatedResource func(target interface{}) error
}

type GenericReplicator struct {
	ReplicatorConfig
	Store      cache.Store
	Controller cache.Controller
	Context    context.Context

	UpdateFuncs UpdateFuncs

	// ReplicateToList is a set that caches the names of all resources that have a
	// "replicate-to" annotation.
	ReplicateToList map[string]struct{}

	// ReplicateToMatchingList is a set that caches the names of all resources
	// that have a "replicate-to-matching" annotation.
	ReplicateToMatchingList map[string]labels.Selector
}

// NewGenericReplicator creates a new GenericReplicator
func NewGenericReplicator(ctx context.Context, config ReplicatorConfig) *GenericReplicator {
	repl := GenericReplicator{
		ReplicatorConfig:        config,
		Context:                 ctx,
		ReplicateToList:         make(map[string]struct{}),
		ReplicateToMatchingList: make(map[string]labels.Selector),
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
			UpdateFunc: repl.ResourceUpdated,
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

	for sourceKey := range r.ReplicateToList {
		logger := logger.WithField("resource", sourceKey)
		obj, err := r.ObjectFromStore(sourceKey)
		if err != nil {
			log.WithError(err).Error("error fetching object from store")
			continue
		}

		objectMeta := MustGetObject(obj)
		namespacePatterns, found := objectMeta.GetAnnotations()[ReplicateTo]
		if found {
			if err := r.replicateResourceToMatchingNamespaces(obj, namespacePatterns, []v1.Namespace{*ns}); err != nil {
				logger.
					WithError(err).
					Errorf("Failed replicating the resource to the new namespace: %s", ns.Name)
			}
		}
	}

	namespaceLabels := labels.Set(ns.Labels)
	for sourceKey, selector := range r.ReplicateToMatchingList {
		logger := logger.WithField("resource", sourceKey)

		obj, err := r.ObjectFromStore(sourceKey)
		if err != nil {
			log.WithError(err).Error("error fetching object from store")
			continue
		}

		if !selector.Matches(namespaceLabels) {
			continue
		}

		if _, err := r.replicateResourceToNamespaces(obj, []v1.Namespace{*ns}); err != nil {
			logger.WithError(err).Error("error while replicating object to namespace")
		}
	}
}

// NamespaceUpdated checks if namespace's labels changed and deletes any 'replicate-to-matching' resources
// the namespace no longer qualifies for. Then it attempts to replicate resources into the updated ns based
// on the updated set of labels
func (r *GenericReplicator) NamespaceUpdated(nsOld *v1.Namespace, nsNew *v1.Namespace) {
	logger := log.WithField("kind", r.Kind).WithField("target", nsNew.Name)

	if reflect.DeepEqual(nsNew.Labels, nsOld.Labels) {
		logger.Debug("labels did not change")
		return
	}

	logger.Infof("labels of namespace %s changed, attempting to delete %ses that no longer match", nsNew.Name, strings.TrimSuffix(r.Kind, "e"))
	// delete any resources where namespace labels no longer match
	var newLabelSet labels.Set = nsNew.Labels
	var oldLabelSet labels.Set = nsOld.Labels
	// check 'replicate-to-matching' resources against new labels
	for sourceKey, selector := range r.ReplicateToMatchingList {
		if selector.Matches(oldLabelSet) && !selector.Matches(newLabelSet) {
			obj, err := r.ObjectFromStore(sourceKey)
			if err != nil {
				log.WithError(err).Error("error fetching object from store")
				continue
			}
			// delete resource from the updated namespace
			logger.Infof("removed %s %s from %s", r.Kind, sourceKey, nsNew.Name)
			r.DeleteResourceInNamespaces(obj, []v1.Namespace{*nsNew})
		}
	}

	// replicate resources to updated ns
	logger.Infof("labels of namespace %s changed, attempting to replicate %ses", nsNew.Name, strings.TrimSuffix(r.Kind, "e"))
	r.NamespaceAdded(nsNew)

}

// ResourceAdded checks resources with ReplicateTo or ReplicateFromAnnotation annotation
func (r *GenericReplicator) ResourceAdded(obj interface{}) {
	objectMeta := MustGetObject(obj)
	sourceKey := MustGetKey(objectMeta)
	logger := log.WithField("kind", r.Kind).WithField("resource", sourceKey)

	if IsManagedBy(MustGetObject(objectMeta)) {
		return
	}

	annotations := objectMeta.GetAnnotations()

	// Match resources with "replicate-to" annotation
	if namespacePatterns, ok := annotations[ReplicateTo]; ok {
		r.ReplicateToList[sourceKey] = struct{}{}

		namespaces, _ := r.ListNamespaces()
		if err := r.replicateResourceToMatchingNamespaces(obj, namespacePatterns, namespaces); err != nil {
			logger.WithError(err).Errorf("could not replicate object to other namespaces")
		}
	} else {
		delete(r.ReplicateToList, sourceKey)
	}

	// Match resources with "replicate-to-matching" annotations
	if namespaceSelectorString, ok := annotations[ReplicateToMatching]; ok {
		namespaceSelector, err := labels.Parse(namespaceSelectorString)
		if err != nil {
			delete(r.ReplicateToMatchingList, sourceKey)
			logger.WithError(err).Error("failed to parse label selector")
			return
		}

		r.ReplicateToMatchingList[sourceKey] = namespaceSelector
		if err := r.replicateResourceToMatchingNamespacesByLabel(obj, namespaceSelector); err != nil {
			logger.WithError(err).Error("error while replicating by label selector")
		}
	} else {
		delete(r.ReplicateToMatchingList, sourceKey)
	}
}

// ResourceUpdated checks resources with ReplicateTo or ReplicateFromAnnotation annotation
func (r *GenericReplicator) ResourceUpdated(old interface{}, new interface{}) {
	oldAnnotations := MustGetObject(old).GetAnnotations()
	newAnnotations := MustGetObject(new).GetAnnotations()

	if !reflect.DeepEqual(oldAnnotations, newAnnotations) {
		r.deleteOldReplicateToResources(old, new)
		r.deleteOldReplicateToMatchingResources(old, new)
	}

	r.ResourceAdded(new)
}

// replicateResourceToMatchingNamespaces replicates resources with ReplicateTo annotation
func (r *GenericReplicator) replicateResourceToMatchingNamespaces(obj interface{}, patterns string, namespaceList []v1.Namespace) error {
	cacheKey := MustGetKey(obj)

	replicateTo := r.getFilteredNamespaces(MustGetObject(obj).GetNamespace(), patterns, namespaceList)

	if replicated, err := r.replicateResourceToNamespaces(obj, replicateTo); err != nil {
		return errors.Wrapf(err, "Replicated %s to %d out of %d namespaces",
			cacheKey, len(replicated), len(replicateTo),
		)
	}

	return nil
}

// replicateResourceToNamespaces will replicate the given object into target namespaces. It will return a list of
// Namespaces it was successful in replicating into
func (r *GenericReplicator) replicateResourceToNamespaces(obj interface{}, targets []v1.Namespace) (replicatedTo []v1.Namespace, err error) {
	cacheKey := MustGetKey(obj)

	for _, namespace := range targets {
		if innerErr := r.UpdateFuncs.ReplicateObjectTo(obj, &namespace); innerErr != nil {
			err = multierror.Append(err, errors.Wrapf(innerErr, "Failed to replicate %s %s -> %s: %v",
				r.Kind, cacheKey, namespace.Name, innerErr,
			))
		} else {
			replicatedTo = append(replicatedTo, namespace)
		}
	}

	return
}

// replicateResourceToMatchingNamespacesByLabel replicates to resources in namespaces with selected labels
func (r *GenericReplicator) replicateResourceToMatchingNamespacesByLabel(obj interface{}, selector labels.Selector) error {
	cacheKey := MustGetKey(obj)

	namespaces, err := r.ListNamespaces(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return errors.Wrap(err, "error while listing namespaces by selector")
	}

	if replicated, err := r.replicateResourceToNamespaces(obj, namespaces); err != nil {
		return errors.Wrapf(err, "Replicated %s to %d out of %d namespaces",
			cacheKey, len(replicated), len(namespaces),
		)
	}

	return nil
}

// ObjectFromStore gets object from store cache
func (r *GenericReplicator) ObjectFromStore(key string) (interface{}, error) {
	obj, exists, err := r.Store.GetByKey(key)
	if err != nil {
		return nil, errors.Errorf("could not get %s %s: %s", r.Kind, key, err)
	}

	if !exists {
		return nil, errors.Errorf("could not get %s %s: does not exist", r.Kind, key)
	}

	return obj, nil
}

// ResourceDeleted watches for the deletion of resources
func (r *GenericReplicator) ResourceDeleted(source interface{}) {
	if IsManagedBy(MustGetObject(source)) {
		return
	}

	sourceKey := MustGetKey(source)
	logger := log.WithField("kind", r.Kind).WithField("source", sourceKey)
	logger.Debugf("Deleting dependents of %s %s", r.Kind, sourceKey)

	r.ResourceDeletedReplicateTo(source)

	delete(r.ReplicateToList, sourceKey)
}

// ResourceDeletedReplicateTo deletes dependent resources that were replicated to
func (r *GenericReplicator) ResourceDeletedReplicateTo(source interface{}) {
	sourceKey := MustGetKey(source)
	logger := log.WithField("kind", r.Kind).WithField("source", sourceKey)

	if err := r.DeleteFilteredNamespaceResources(source); err != nil {
		logger.WithError(err).Errorf("Could not delete resources from filtered namespaces")
	}

	if err := r.DeletedLabelSelectedNamespaceResources(source); err != nil {
		logger.WithError(err).Errorf("Could not delete resources from label selected namespaces")
	}
}

// DeleteFilteredNamespaceResources deletes resources from a filtered namespace list
func (r *GenericReplicator) DeleteFilteredNamespaceResources(source interface{}) error {
	objMeta := MustGetObject(source)

	patterns, replicateTo := objMeta.GetAnnotations()[ReplicateTo]
	if !replicateTo {
		return nil
	}

	namespaces, err := r.ListFilteredNamespaces(objMeta.GetNamespace(), patterns)
	if err != nil {
		return errors.Wrapf(err, "Failed to list namespaces: %v", err)
	}

	r.DeleteResourceInNamespaces(source, namespaces)
	return nil
}

// DeletedLabelSelectedNamespaceResources deletes resources from label selected namespaces
func (r *GenericReplicator) DeletedLabelSelectedNamespaceResources(source interface{}) error {
	objMeta := MustGetObject(source)

	namespaceSelectorString, replicateToMatching := objMeta.GetAnnotations()[ReplicateToMatching]
	if replicateToMatching {
		return nil
	}

	namespaces, err := r.ListLabelSelectedNamespaces(namespaceSelectorString)
	if err != nil {
		return err
	}

	r.DeleteResourceInNamespaces(source, namespaces)
	return nil
}

// DeleteResourceInNamespaces deletes resources in a list of namespaces acquired by evaluating namespace labels
func (r *GenericReplicator) DeleteResourceInNamespaces(source interface{}, namespaces []v1.Namespace) {
	for _, namespace := range namespaces {
		r.DeleteResource(namespace, source)
	}
}

// DeleteResource deletes a single resource from the provided namespace
func (r *GenericReplicator) DeleteResource(namespace v1.Namespace, source interface{}) {
	sourceKey := MustGetKey(source)
	logger := log.WithField("kind", r.Kind).WithField("source", sourceKey)
	objMeta := MustGetObject(source)

	if namespace.Name == objMeta.GetNamespace() {
		return
	}
	targetLocation := fmt.Sprintf("%s/%s", namespace.Name, objMeta.GetName())
	targetResource, err := r.ObjectFromStore(targetLocation)
	if err != nil {
		logger.WithError(err).Errorf("Could not get objectMeta %s: %+v", targetLocation, err)
		return
	}

	logger.Infof("Deleting %s: %s", r.Kind, targetLocation)
	if err := r.UpdateFuncs.DeleteReplicatedResource(targetResource); err != nil {
		logger.WithError(err).Errorf("Could not delete resource %s: %+v", targetLocation, err)
	}
	r.Store.Delete(source)
}

func (r *GenericReplicator) deleteOldReplicateToResources(old, new interface{}) {
	oldObj := MustGetObject(old)

	oldPatterns, oldReplicateTo := oldObj.GetAnnotations()[ReplicateTo]
	if !oldReplicateTo {
		return
	}

	oldNamespaces, err := r.ListFilteredNamespaces(oldObj.GetNamespace(), oldPatterns)
	if err != nil || len(oldNamespaces) == 0 {
		return
	}

	newObj := MustGetObject(new)

	newPatterns, newReplicateTo := newObj.GetAnnotations()[ReplicateTo]
	if !newReplicateTo {
		r.DeleteResourceInNamespaces(old, oldNamespaces)
		return
	}

	newNamespaces, err := r.ListFilteredNamespaces(newObj.GetNamespace(), newPatterns)
	if err != nil || len(newNamespaces) == 0 {
		r.DeleteResourceInNamespaces(old, oldNamespaces)
		return
	}

	var removedNamespaces []v1.Namespace
Outer:
	for _, oldNamespace := range oldNamespaces {
		for _, newNamespace := range newNamespaces {
			log.Debugf("%s == %s", oldNamespace.Name, newNamespace.Name)
			if oldNamespace.Name == newNamespace.Name {
				continue Outer
			}
		}
		removedNamespaces = append(removedNamespaces, oldNamespace)
	}

	r.DeleteResourceInNamespaces(old, removedNamespaces)
}

func (r *GenericReplicator) deleteOldReplicateToMatchingResources(old, new interface{}) {

}

// ListNamespaces is a simple wrapper for listing namespaces
func (r *GenericReplicator) ListNamespaces(listOptions ...metav1.ListOptions) ([]v1.Namespace, error) {
	if len(listOptions) == 0 {
		listOptions = append(listOptions, metav1.ListOptions{})
	}
	namespaceList, err := r.Client.CoreV1().Namespaces().List(r.Context, listOptions[0])
	return namespaceList.Items, err
}

// ListLabelSelectedNamespaces retrieves list of namespaces that meet the label selector
func (r *GenericReplicator) ListLabelSelectedNamespaces(namespaceSelectorString string) ([]v1.Namespace, error) {
	namespaceSelector, err := labels.Parse(namespaceSelectorString)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed parse namespace selector: %v", err)
	}

	namespaces, err := r.ListNamespaces(metav1.ListOptions{LabelSelector: namespaceSelector.String()})
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to list namespaces: %v", err)
	}

	return namespaces, nil
}

// ListFilteredNamespaces retrieve list of namespaces that match the provided patterns
func (r *GenericReplicator) ListFilteredNamespaces(current string, patterns string) ([]v1.Namespace, error) {
	namespaces, err := r.ListNamespaces()
	if err != nil {
		return nil, err
	}

	return r.getFilteredNamespaces(current, patterns, namespaces), nil
}

// getFilteredNamespaces will check the provided filters return a list of namespace that fit the filter conditions
func (r *GenericReplicator) getFilteredNamespaces(current string, patterns string, namespaces []v1.Namespace) []v1.Namespace {

	filtered := make([]v1.Namespace, 0)
	for _, namespace := range namespaces {
		if namespace.Name == current {
			continue
		}
		for _, ns := range StringToPatternList(patterns) {
			if matched := ns.MatchString(namespace.Name); matched {
				filtered = append(filtered, namespace)
				break
			}
		}
	}
	return filtered
}
