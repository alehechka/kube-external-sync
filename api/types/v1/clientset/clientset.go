package clientset

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

// KubeExternalSyncClientset represents the REST client for kube-external-sync
type KubeExternalSyncClientset struct {
	client rest.Interface
}

// NewForConfig creates a REST Client for the kube-secret-sync CustomResourceDefinitions
func NewForConfig(c *rest.Config) (*KubeExternalSyncClientset, error) {
	AddToScheme(scheme.Scheme)

	config := *c
	config.ContentConfig.GroupVersion = &schema.GroupVersion{Group: GroupName, Version: GroupVersion}
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	config.UserAgent = rest.DefaultKubernetesUserAgent()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	return &KubeExternalSyncClientset{client: client}, nil
}

func (c *KubeExternalSyncClientset) ExternalSyncRules() ExternalSyncRuleInterface {
	return newExternalSyncRules(c)
}
