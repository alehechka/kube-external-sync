package v1

import (
	"context"
	"fmt"

	"github.com/alehechka/kube-external-sync/api/types"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// +kubebuilder:object:root=true

// ExternalSyncRule is the definition for the ExternalSyncRule CRD
type ExternalSyncRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ExternalSyncRuleSpec `json:"spec"`
}

// +kubebuilder:object:root=true

// ExternalSyncRuleList is the definition for the ExternalSyncRule CRD list
type ExternalSyncRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ExternalSyncRule `json:"items"`
}

// +kubebuilder:object:generate=true

// ExternalSyncRuleSpec is the spec attribute of the ExternalSyncRule CRD
type ExternalSyncRuleSpec struct {
	Namespace string   `json:"namespace"`
	Service   *Service `json:"service,omitempty"`
	Ingress   *Ingress `json:"ingress,omitempty"`
	Rules     Rules    `json:"rules"`
}

// +kubebuilder:object:generate=true

// Service defines the attributes of the Service to sync
type Service struct {
	Name               string `json:"name"`
	Kind               string `json:"kind"`
	ExternalNameSuffix string `json:"externalNameSuffix"`
}

func (service *Service) IsService() bool {
	return service.Kind == "Service"
}

func (service *Service) IsTraefikService() bool {
	return service.Kind == "Traefik Service"
}

// +kubebuilder:object:generate=true

// Ingress defines the attributes of the Ingress to sync
type Ingress struct {
	Name           string `json:"name"`
	Kind           string `json:"kind"`
	TopLevelDomain string `json:"topLevelDomain"`
}

// +kubebuilder:object:generate=true

// Rules contains all rules for the secret to follow
type Rules struct {
	Namespaces NamespaceRules `json:"namespaces"`
}

// +kubebuilder:object:generate=true

// NamespaceRules include all rules for namepsaces to sync to.
type NamespaceRules struct {
	Exclude           types.StringSlice `json:"exclude"`
	ExcludeRegex      types.StringSlice `json:"excludeRegex"`
	Include           types.StringSlice `json:"include"`
	IncludeRegex      types.StringSlice `json:"includeRegex"`
	IncludeAnnotation types.Annotations `json:"includeAnnotation"`
}

func (rule *ExternalSyncRule) HasService() bool {
	return rule.Spec.Service != nil &&
		len(rule.Spec.Service.Name) > 0
}

func (rule *ExternalSyncRule) HasIngress() bool {
	return rule.Spec.Ingress != nil &&
		len(rule.Spec.Ingress.Name) > 0
}

func (rule *ExternalSyncRule) ShouldSyncService(service *v1.Service) bool {
	return rule.HasService() &&
		rule.Spec.Namespace == service.Namespace &&
		rule.Spec.Service.Name == service.Name
}

// ShouldSyncNamespace determines whether or not the given Namespace should be synced
func (rule *ExternalSyncRule) ShouldSyncNamespace(namespace *v1.Namespace) bool {
	rules := rule.Spec.Rules

	if rule.Spec.Namespace == namespace.Name {
		return false
	}

	if rules.Namespaces.Exclude.IsExcluded(namespace.Name) || rules.Namespaces.ExcludeRegex.IsRegexExcluded(namespace.Name) {
		return false
	}

	if rules.Namespaces.Include.IsEmpty() && rules.Namespaces.IncludeRegex.IsEmpty() && rules.Namespaces.IncludeAnnotation.IsEmpty() {
		return true
	}

	if rules.Namespaces.Include.IsIncluded(namespace.Name) || rules.Namespaces.IncludeRegex.IsRegexIncluded(namespace.Name) || rules.Namespaces.IncludeAnnotation.AnnotatesNamespace(namespace) {
		return true
	}

	return false
}

// Namespaces returns a list of all namespaces that the given Rule allows for syncing
func (rule *ExternalSyncRule) Namespaces(ctx context.Context, clientset kubernetes.Interface) (namespaces []v1.Namespace) {
	list, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return
	}

	for _, namespace := range list.Items {
		if rule.ShouldSyncNamespace(&namespace) {
			namespaces = append(namespaces, namespace)
		}
	}

	return
}

// ShouldSyncNamespace iterates over the list to determine whether or not the given Namespace should be synced
func (list *ExternalSyncRuleList) ShouldSyncNamespace(namespace *v1.Namespace) bool {
	for _, rule := range list.Items {
		if rule.ShouldSyncNamespace(namespace) {
			return true
		}
	}

	return false
}

func (rule *ExternalSyncRule) ServiceExternalName() string {
	if rule == nil || !rule.HasService() {
		return ""
	}

	return fmt.Sprintf("%s.%s.%s", rule.Spec.Service.Name, rule.Spec.Namespace, rule.Spec.Service.ExternalNameSuffix)
}
