package service

import (
	"context"
	"fmt"
	"time"

	"github.com/alehechka/kube-external-sync/client/replicate/common"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

type Replicator struct {
	*common.GenericReplicator
}

// NewReplicator creates a new service replicator
func NewReplicator(ctx context.Context, client kubernetes.Interface, resyncPeriod time.Duration) common.Replicator {
	repl := Replicator{
		GenericReplicator: common.NewGenericReplicator(ctx, common.ReplicatorConfig{
			Kind:         "Service",
			ObjType:      &v1.Service{},
			ResyncPeriod: resyncPeriod,
			Client:       client,
			ListFunc: func(lo metav1.ListOptions) (runtime.Object, error) {
				return client.CoreV1().Services(v1.NamespaceAll).List(ctx, lo)
			},
			WatchFunc: func(lo metav1.ListOptions) (watch.Interface, error) {
				return client.CoreV1().Services(v1.NamespaceAll).Watch(ctx, lo)
			},
		}),
	}
	repl.UpdateFuncs = common.UpdateFuncs{
		ReplicateDataFrom:        repl.ReplicateDataFrom,
		ReplicateObjectTo:        repl.ReplicateObjectTo,
		DeleteReplicatedResource: repl.DeleteReplicatedResource,
	}

	return &repl
}

// ReplicateDataFrom takes a source object and copies over data to target object
func (r *Replicator) ReplicateDataFrom(sourceObj interface{}, targetObj interface{}) error {
	source := sourceObj.(*v1.Service)
	target := targetObj.(*v1.Service)

	logger := log.
		WithField("kind", r.Kind).
		WithField("source", common.MustGetKey(source)).
		WithField("target", common.MustGetKey(target))

	if !common.IsManagedBy(target) {
		logger.Debugf("target is not managed and will not be synced")
		return nil
	}

	targetVersion, ok := target.Annotations[common.ReplicatedFromVersionAnnotation]
	sourceVersion := source.ResourceVersion

	if ok && targetVersion == sourceVersion {
		logger.Debugf("target is already up-to-date")
		return nil
	}

	prepared := prepareExternalNameService(target.Namespace, source)
	service, err := r.Client.CoreV1().Services(target.Namespace).Update(r.Context, prepared, metav1.UpdateOptions{})
	if err != nil {
		err = errors.Wrapf(err, "Failed updating target %s", common.MustGetKey(prepared))
	} else if err = r.Store.Update(service); err != nil {
		err = errors.Wrapf(err, "Failed to update cache for %s: %v", common.MustGetKey(prepared), err)
	}
	return err
}

// ReplicateObjectTo copies the whole object to target namespace
func (r *Replicator) ReplicateObjectTo(sourceObj interface{}, targetNamespace *v1.Namespace) error {
	source := sourceObj.(*v1.Service)
	targetLocation := fmt.Sprintf("%s/%s", targetNamespace.Name, source.Name)

	targetResource, exists, err := r.Store.GetByKey(targetLocation)
	if err != nil {
		return errors.Wrapf(err, "Could not get %s from cache!", targetLocation)
	}

	if exists {
		return r.ReplicateDataFrom(source, (targetResource).(*v1.Service))
	}

	prepared := prepareExternalNameService(targetNamespace.Name, source)
	service, err := r.Client.CoreV1().Services(targetNamespace.Name).Create(r.Context, prepared, metav1.CreateOptions{})
	if err != nil {
		err = errors.Wrapf(err, "Failed creating target %s", common.MustGetKey(prepared))
	} else if err = r.Store.Update(service); err != nil {
		err = errors.Wrapf(err, "Failed to update cache for %s: %v", common.MustGetKey(prepared), err)
	}
	return err
}

// DeleteReplicatedResource deletes a resource replicated by ReplicateTo annotation
func (r *Replicator) DeleteReplicatedResource(targetResource interface{}) error {
	service := targetResource.(*v1.Service)
	return r.Client.CoreV1().Services(service.Namespace).Delete(r.Context, service.Name, metav1.DeleteOptions{})
}

func prepareExternalNameService(namespace string, source *v1.Service) *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            source.Name,
			Namespace:       namespace,
			Labels:          common.PrepareLabels(source.ObjectMeta),
			Annotations:     common.PrepareAnnotations(source.ObjectMeta),
			OwnerReferences: common.PrepareOwnerReferences(source.ObjectMeta),
		},
		Spec: v1.ServiceSpec{
			Type:         v1.ServiceTypeExternalName,
			ExternalName: prepareExternalName(source.Namespace, source),
			Ports:        source.Spec.Ports,
		},
	}
}

func prepareExternalName(namespace string, source *v1.Service) string {
	return fmt.Sprintf("%s.%s.%s", source.Name, namespace, getExternalNameSuffix(source))
}

func getExternalNameSuffix(source *v1.Service) string {
	if suffix, ok := source.Annotations[common.ExternalNameSuffix]; ok && len(suffix) > 0 {
		return suffix
	}

	return common.DefaultExternalNameSuffix
}
