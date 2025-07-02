package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Image",type="string",JSONPath=".spec.image"
//+kubebuilder:printcolumn:name="Namespace",type="string",JSONPath=".spec.namespace"
//+kubebuilder:printcolumn:name="Secrets Found",type="integer",JSONPath=".spec.secretsFound"
//+kubebuilder:printcolumn:name="Timestamp",type="date",JSONPath=".spec.timestamp"
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status"

// SecretScanning is the Schema for the secretscannings API
type SecretScanning struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SecretScanningSpec   `json:"spec,omitempty"`
	Status SecretScanningStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SecretScanningList contains a list of SecretScanning
type SecretScanningList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecretScanning `json:"items"`
}

// SecretScanningSpec defines the desired state of SecretScanning
type SecretScanningSpec struct {
	Image            string `json:"image"`
	Namespace        string `json:"namespace"`
	Timestamp        string `json:"timestamp"`
	ClusterName      string `json:"clusterName"`
	JobID            string `json:"jobID"`
	TrufflehogReport string `json:"trufflehogReport"`     // Complete trufflehog JSON report
	SecretsFound     int    `json:"secretsFound"`         // Number of secrets found
	ReportPath       string `json:"reportPath,omitempty"` // Path to the local report files
}

// SecretScanningStatus defines the observed state of SecretScanning
type SecretScanningStatus struct {
	Status      string `json:"status,omitempty"`
	LastUpdated string `json:"lastUpdated,omitempty"`
}
