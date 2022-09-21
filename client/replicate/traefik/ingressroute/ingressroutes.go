package ingressroute

import (
	"context"
	"fmt"
	"time"

	"github.com/alehechka/kube-external-sync/client/replicate/common"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/generated/clientset/versioned"
	"github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	"github.com/traefik/traefik/v2/pkg/types"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

type Replicator struct {
	*common.GenericReplicator
}

// NewReplicator creates a new ingress replicator
func NewReplicator(ctx context.Context, client kubernetes.Interface, traefik *versioned.Clientset, resyncPeriod time.Duration) common.Replicator {

	repl := Replicator{
		GenericReplicator: common.NewGenericReplicator(ctx, common.ReplicatorConfig{
			Kind:          "IngressRoute",
			ObjType:       &v1alpha1.IngressRoute{},
			ResyncPeriod:  resyncPeriod,
			Client:        client,
			TraefikClient: traefik,
			ListFunc: func(lo metav1.ListOptions) (runtime.Object, error) {
				return traefik.TraefikV1alpha1().IngressRoutes(v1.NamespaceAll).List(ctx, lo)
			},
			WatchFunc: func(lo metav1.ListOptions) (watch.Interface, error) {
				return traefik.TraefikV1alpha1().IngressRoutes(v1.NamespaceAll).Watch(ctx, lo)
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
	source := sourceObj.(*v1alpha1.IngressRoute)
	target := targetObj.(*v1alpha1.IngressRoute)

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

	prepared := prepareIngressRoute(target.Namespace, source)
	service, err := r.TraefikClient.TraefikV1alpha1().IngressRoutes(target.Namespace).Update(r.Context, prepared, metav1.UpdateOptions{})
	if err != nil {
		err = errors.Wrapf(err, "Failed updating target %s", common.MustGetKey(prepared))
	} else if err = r.Store.Update(service); err != nil {
		err = errors.Wrapf(err, "Failed to update cache for %s: %v", common.MustGetKey(prepared), err)
	}
	return err
}

// ReplicateObjectTo copies the whole object to target namespace
func (r *Replicator) ReplicateObjectTo(sourceObj interface{}, targetNamespace *v1.Namespace) error {
	source := sourceObj.(*v1alpha1.IngressRoute)
	sourceKey := common.MustGetKey(source)
	targetLocation := fmt.Sprintf("%s/%s", targetNamespace.Name, source.Name)

	logger := log.WithField("source", sourceKey).WithField("target", targetLocation).WithField("kind", r.Kind)
	logger.Infof("Replicating %s to %s", sourceKey, targetNamespace.Name)

	targetResource, err := r.TraefikClient.TraefikV1alpha1().IngressRoutes(targetNamespace.Name).Get(r.Context, source.Name, metav1.GetOptions{})
	if err == nil && targetResource != nil {
		return r.ReplicateDataFrom(source, targetResource)
	}

	prepared := prepareIngressRoute(targetNamespace.Name, source)
	service, err := r.TraefikClient.TraefikV1alpha1().IngressRoutes(targetNamespace.Name).Create(r.Context, prepared, metav1.CreateOptions{})
	if err != nil {
		err = errors.Wrapf(err, "Failed creating target %s", common.MustGetKey(prepared))
	} else if err = r.Store.Update(service); err != nil {
		err = errors.Wrapf(err, "Failed to update cache for %s: %v", common.MustGetKey(prepared), err)
	}
	return err
}

// DeleteReplicatedResource deletes a resource replicated by ReplicateTo annotation
func (r *Replicator) DeleteReplicatedResource(targetResource interface{}) error {
	ingressRoute := targetResource.(*v1alpha1.IngressRoute)

	if !common.IsManagedBy(ingressRoute) {
		log.WithField("kind", r.Kind).WithField("target", common.MustGetKey(ingressRoute)).
			Debugf("target is not managed and will not be deleted")
		return nil
	}

	return r.TraefikClient.TraefikV1alpha1().IngressRoutes(ingressRoute.Namespace).Delete(r.Context, ingressRoute.Name, metav1.DeleteOptions{})
}

func prepareIngressRoute(namespace string, source *v1alpha1.IngressRoute) *v1alpha1.IngressRoute {
	return &v1alpha1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:            source.Name,
			Namespace:       namespace,
			Labels:          common.PrepareLabels(source.ObjectMeta),
			Annotations:     common.PrepareAnnotations(source.ObjectMeta),
			OwnerReferences: common.PrepareOwnerReferences(source.ObjectMeta),
		},
		Spec: v1alpha1.IngressRouteSpec{
			EntryPoints: source.Spec.EntryPoints,
			Routes:      prepareRoutes(namespace, source),
			TLS:         prepareTLS(namespace, source),
		},
	}
}

func prepareRoutes(namespace string, source *v1alpha1.IngressRoute) (routes []v1alpha1.Route) {
	for _, route := range source.Spec.Routes {
		newRoute := v1alpha1.Route{
			Kind:        route.Kind,
			Middlewares: route.Middlewares,
			Services:    route.Services,
			Priority:    route.Priority,
		}

		newRoute.Match = common.PrepareRouteMatch(namespace, route.Match)

		routes = append(routes, newRoute)
	}

	return
}

func prepareTLS(namespace string, source *v1alpha1.IngressRoute) *v1alpha1.TLS {
	annotations := source.GetAnnotations()

	if tld, ok := annotations[common.TopLevelDomain]; ok {
		return &v1alpha1.TLS{
			SecretName:   annotations[common.TLDSecretName],
			Options:      source.Spec.TLS.Options,
			Store:        source.Spec.TLS.Store,
			CertResolver: source.Spec.TLS.CertResolver,
			Domains: []types.Domain{{
				Main: common.PrepareTLD(namespace, tld),
			}},
		}
	}

	return &v1alpha1.TLS{
		SecretName:   source.Spec.TLS.SecretName,
		Options:      source.Spec.TLS.Options,
		Store:        source.Spec.TLS.Store,
		CertResolver: source.Spec.TLS.CertResolver,
		Domains:      prepareDomains(namespace, source),
	}
}

func prepareDomains(namespace string, source *v1alpha1.IngressRoute) (domains []types.Domain) {
	for _, domain := range source.Spec.TLS.Domains {
		newDomain := types.Domain{Main: common.PrepareTLD(namespace, domain.Main)}
		for _, san := range domain.SANs {
			newDomain.SANs = append(newDomain.SANs, common.PrepareTLD(namespace, san))
		}
		domains = append(domains, newDomain)
	}
	return
}
