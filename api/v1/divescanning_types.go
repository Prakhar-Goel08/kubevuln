package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Image",type="string",JSONPath=".spec.image"
//+kubebuilder:printcolumn:name="Namespace",type="string",JSONPath=".spec.namespace"
//+kubebuilder:printcolumn:name="Image Size",type="integer",JSONPath=".spec.imageSize"
//+kubebuilder:printcolumn:name="Layers",type="integer",JSONPath=".spec.layers"
//+kubebuilder:printcolumn:name="Efficiency",type="number",JSONPath=".spec.efficiencyScore"
//+kubebuilder:printcolumn:name="Timestamp",type="date",JSONPath=".spec.timestamp"
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status"

// DiveScanning is the Schema for the divescannings API
type DiveScanning struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DiveScanningSpec   `json:"spec,omitempty"`
	Status DiveScanningStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DiveScanningList contains a list of DiveScanning
type DiveScanningList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DiveScanning `json:"items"`
}

// DiveScanningSpec defines the desired state of DiveScanning
type DiveScanningSpec struct {
	Image           string  `json:"image"`
	Namespace       string  `json:"namespace"`
	Timestamp       string  `json:"timestamp"`
	ClusterName     string  `json:"clusterName"`
	JobID           string  `json:"jobID"`
	DiveReport      string  `json:"diveReport"`                // Complete dive JSON report
	ImageSize       int64   `json:"imageSize,omitempty"`       // Size of the analyzed image in bytes
	Layers          int     `json:"layers,omitempty"`          // Number of layers in the image
	EfficiencyScore float64 `json:"efficiencyScore,omitempty"` // Dive efficiency score
	ReportPath      string  `json:"reportPath,omitempty"`      // Path to the local report files
}

// DiveScanningStatus defines the observed state of DiveScanning
type DiveScanningStatus struct {
	Status      string `json:"status,omitempty"`
	LastUpdated string `json:"lastUpdated,omitempty"`
}
