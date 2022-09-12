package clientset

import (
	"context"
	"time"

	typesv1 "github.com/alehechka/kube-external-sync/api/types/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

const resource = "externalsyncrules"

// ExternalSyncRuleGetter has a method to return a ExternalSyncRuleInterface.
type ExternalSyncRuleGetter interface {
	ExternalSyncRules() ExternalSyncRuleInterface
}

// externalSyncRules implements ExternalSyncRuleInterface
type externalSyncRules struct {
	client rest.Interface
}

// newExternalSyncRules returns a ExternalSyncRules
func newExternalSyncRules(c *KubeExternalSyncClientset) *externalSyncRules {
	return &externalSyncRules{
		client: c.client,
	}
}

// ExternalSyncRuleInterface has methods to work with ExternalSyncRule resources.
type ExternalSyncRuleInterface interface {
	List(ctx context.Context, opts metav1.ListOptions) (*typesv1.ExternalSyncRuleList, error)
	Get(ctx context.Context, name string, options metav1.GetOptions) (*typesv1.ExternalSyncRule, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
}

func (c *externalSyncRules) List(ctx context.Context, opts metav1.ListOptions) (*typesv1.ExternalSyncRuleList, error) {
	result := typesv1.ExternalSyncRuleList{}
	err := c.client.
		Get().
		Resource(resource).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (c *externalSyncRules) Get(ctx context.Context, name string, opts metav1.GetOptions) (*typesv1.ExternalSyncRule, error) {
	result := typesv1.ExternalSyncRule{}
	err := c.client.
		Get().
		Resource(resource).
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (c *externalSyncRules) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.
		Get().
		Resource(resource).
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}
