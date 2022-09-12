package v1

import (
	"github.com/alehechka/kube-external-sync/api/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	Service *Service `json:"service,omitempty"`
	Ingress *Ingress `json:"ingress,omitempty"`
	Rules   Rules    `json:"rules"`
}

// +kubebuilder:object:generate=true

// Service defines the attributes of the Service to sync
type Service struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Kind      string `json:"kind"`
}

// +kubebuilder:object:generate=true

// Ingress defines the attributes of the Ingress to sync
type Ingress struct {
	Name           string `json:"name"`
	Namespace      string `json:"namespace"`
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
		len(rule.Spec.Service.Name) > 0 &&
		len(rule.Spec.Service.Namespace) > 0
}

func (rule *ExternalSyncRule) HasIngress() bool {
	return rule.Spec.Ingress != nil &&
		len(rule.Spec.Ingress.Name) > 0 &&
		len(rule.Spec.Ingress.Namespace) > 0
}
