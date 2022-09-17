package ingress

import (
	"context"
	"fmt"
	"time"

	"github.com/alehechka/kube-external-sync/client/replicate/common"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

type Replicator struct {
	*common.GenericReplicator
}

// NewReplicator creates a new ingress replicator
func NewReplicator(ctx context.Context, client kubernetes.Interface, resyncPeriod time.Duration) common.Replicator {
	repl := Replicator{
		GenericReplicator: common.NewGenericReplicator(ctx, common.ReplicatorConfig{
			Kind:         "Ingress",
			ObjType:      &networkingv1.Ingress{},
			ResyncPeriod: resyncPeriod,
			Client:       client,
			ListFunc: func(lo metav1.ListOptions) (runtime.Object, error) {
				return client.NetworkingV1().Ingresses(v1.NamespaceAll).List(ctx, lo)
			},
			WatchFunc: func(lo metav1.ListOptions) (watch.Interface, error) {
				return client.NetworkingV1().Ingresses(v1.NamespaceAll).Watch(ctx, lo)
			},
		}),
	}
	repl.UpdateFuncs = common.UpdateFuncs{
		ReplicateDataFrom: repl.ReplicateDataFrom,
		ReplicateObjectTo: repl.ReplicateObjectTo,
	}

	return &repl
}

// ReplicateDataFrom takes a source object and copies over data to target object
func (r *Replicator) ReplicateDataFrom(sourceObj interface{}, targetObj interface{}) error {
	source := sourceObj.(*networkingv1.Ingress)
	target := targetObj.(*networkingv1.Ingress)

	logger := log.
		WithField("kind", r.Kind).
		WithField("source", common.MustGetKey(source)).
		WithField("tartget", common.MustGetKey(target))

	logger.Infof("ReplicateDataFrom")

	return nil
}

// ReplicateObjectTo copies the whole object to target namespace
func (r *Replicator) ReplicateObjectTo(sourceObj interface{}, target *v1.Namespace) error {
	source := sourceObj.(*networkingv1.Ingress)
	targetLocation := fmt.Sprintf("%s/%s", target.Name, source.Name)

	logger := log.
		WithField("kind", r.Kind).
		WithField("source", common.MustGetKey(source)).
		WithField("target", targetLocation)

	logger.Infof("ReplicateObjectTo")

	return nil
}
