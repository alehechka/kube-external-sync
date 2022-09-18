package ingress

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/alehechka/kube-external-sync/client/replicate/common"
	"github.com/pkg/errors"
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
		ReplicateDataFrom:        repl.ReplicateDataFrom,
		ReplicateObjectTo:        repl.ReplicateObjectTo,
		DeleteReplicatedResource: repl.DeleteReplicatedResource,
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

	prepared := prepareIngress(target.Namespace, source)
	service, err := r.Client.NetworkingV1().Ingresses(target.Namespace).Update(r.Context, prepared, metav1.UpdateOptions{})
	if err != nil {
		err = errors.Wrapf(err, "Failed updating target %s", common.MustGetKey(prepared))
	} else if err = r.Store.Update(service); err != nil {
		err = errors.Wrapf(err, "Failed to update cache for %s: %v", common.MustGetKey(prepared), err)
	}
	return err
}

// ReplicateObjectTo copies the whole object to target namespace
func (r *Replicator) ReplicateObjectTo(sourceObj interface{}, targetNamespace *v1.Namespace) error {
	source := sourceObj.(*networkingv1.Ingress)
	targetLocation := fmt.Sprintf("%s/%s", targetNamespace.Name, source.Name)

	targetResource, exists, err := r.Store.GetByKey(targetLocation)
	if err != nil {
		return errors.Wrapf(err, "Could not get %s from cache!", targetLocation)
	}

	if exists {
		return r.ReplicateDataFrom(source, (targetResource).(*networkingv1.Ingress))
	}

	prepared := prepareIngress(targetNamespace.Name, source)
	service, err := r.Client.NetworkingV1().Ingresses(targetNamespace.Name).Create(r.Context, prepared, metav1.CreateOptions{})
	if err != nil {
		err = errors.Wrapf(err, "Failed creating target %s", common.MustGetKey(prepared))
	} else if err = r.Store.Update(service); err != nil {
		err = errors.Wrapf(err, "Failed to update cache for %s: %v", common.MustGetKey(prepared), err)
	}
	return err
}

// DeleteReplicatedResource deletes a resource replicated by ReplicateTo annotation
func (r *Replicator) DeleteReplicatedResource(targetResource interface{}) error {
	ingress := targetResource.(*networkingv1.Ingress)
	return r.Client.NetworkingV1().Ingresses(ingress.Namespace).Delete(r.Context, ingress.Name, metav1.DeleteOptions{})
}

func prepareIngress(namespace string, source *networkingv1.Ingress) *networkingv1.Ingress {
	return &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:            source.Name,
			Namespace:       namespace,
			Labels:          common.PrepareLabels(source.ObjectMeta),
			Annotations:     common.PrepareAnnotations(source.ObjectMeta),
			OwnerReferences: common.PrepareOwnerReferences(source.ObjectMeta),
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: source.Spec.IngressClassName,
			DefaultBackend:   source.Spec.DefaultBackend,
			TLS:              prepareTLS(namespace, source),
			Rules:            prepareRules(namespace, source),
		},
	}
}

func prepareTLS(namespace string, source *networkingv1.Ingress) (ingressTLS []networkingv1.IngressTLS) {
	annotations := source.GetAnnotations()

	if tld, ok := annotations[common.TopLevelDomain]; ok {
		return []networkingv1.IngressTLS{{
			SecretName: annotations[common.TLDSecretName],
			Hosts:      []string{prepareTLD(namespace, tld)},
		}}
	}

	for _, tls := range source.Spec.TLS {
		entry := networkingv1.IngressTLS{SecretName: tls.SecretName}
		for _, host := range tls.Hosts {
			entry.Hosts = append(entry.Hosts, prepareTLD(namespace, host))
		}
		ingressTLS = append(ingressTLS, entry)
	}

	return
}

func prepareRules(namespace string, source *networkingv1.Ingress) (rules []networkingv1.IngressRule) {
	tld, ok := source.GetAnnotations()[common.TopLevelDomain]
	prepared := prepareTLD(namespace, tld)

	for _, rule := range source.Spec.Rules {
		host := prepared
		if !ok {
			host = prepareTLD(namespace, rule.Host)
		}

		rules = append(rules, networkingv1.IngressRule{
			Host:             host,
			IngressRuleValue: rule.IngressRuleValue,
		})
	}

	return
}

func prepareTLD(namespace, tld string) string {
	subdomains := strings.Split(tld, ".")
	subdomains[0] = namespace

	return strings.Join(subdomains, ".")
}
