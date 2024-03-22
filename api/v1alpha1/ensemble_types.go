/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"fmt"

	minicluster "github.com/flux-framework/flux-operator/api/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EnsembleSpec defines the desired state of Ensemble
type EnsembleSpec struct {

	// Foo is an example field of Ensemble. Edit ensemble_types.go to remove/update
	Members []Member `json:"members"`

	// Global algorithmt to use, unless a member has a specific algorithm
	// +kubebuilder:default="maintain"
	// +default="maintain"
	//+optional
	Algorithm string `json:"algorithm"`
}

// A member of the ensemble that will run for some number of times,
// optionally with a maximum or minumum
type Member struct {

	// MiniCluster is of a type MiniCluster, the base unit of an ensemble.
	// We do this because we install a flux metrics API within each MiniCluster to manage it
	// TODO where should the user define the size? Here or with the member?
	// +optional
	MiniCluster minicluster.MiniCluster `json:"minicluster,omitempty"`

	// JobSet as the Member
	// +optional
	//JobSet jobset.JobSet `json:"jobset,omitempty"`

	// Job
	// +optional
	//Job batchv1.Job `json:"job,omitempty"`

	// Member specific algorithm to use
	// +kubebuilder:default="maintain"
	// +default="maintain"
	//+optional
	Algorithm string `json:"algorithm"`

	// +kubebuilder:default=1
	// +default=1
	//+optional
	DesiredSize int32 `json:"desiredSize,omitempty"`

	// +kubebuilder:default=1
	// +default=1
	//+optional
	MinSize int32 `json:"minSize,omitempty"`

	// +kubebuilder:default=1
	// +default=1
	//+optional
	MaxSize int32 `json:"maxSize,omitempty"`
}

// EnsembleStatus defines the observed state of Ensemble
type EnsembleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// Validate ensures we have data that is needed, and sets defaults if needed
func (e *Ensemble) Validate() error {
	fmt.Println()

	// Global (entire cluster) settings
	fmt.Printf("🤓 Ensemble.members %d\n", len(e.Spec.Members))

	// If MaxSize is set, it must be greater than size
	if len(e.Spec.Members) < 1 {
		return fmt.Errorf("ensemble must have at least one member")
	}

	for _, member := range e.Spec.Members {
		if member.MinSize > member.MaxSize {
			return fmt.Errorf("ensemble must have at least one member")
		}
		if member.DesiredSize < member.MinSize || member.DesiredSize > member.MaxSize {
			return fmt.Errorf("ensemble desired size must be between min and max size")
		}
	}
	return nil
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Ensemble is the Schema for the ensembles API
type Ensemble struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EnsembleSpec   `json:"spec,omitempty"`
	Status EnsembleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// EnsembleList contains a list of Ensemble
type EnsembleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Ensemble `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Ensemble{}, &EnsembleList{})
}
