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
	"reflect"

	minicluster "github.com/flux-framework/flux-operator/api/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/set"
)

var (
	defaultSidecarbase = "ghcr.io/converged-computing/ensemble-operator-api:rockylinux9"
)

// EnsembleSpec defines the desired state of Ensemble
type EnsembleSpec struct {

	// Foo is an example field of Ensemble. Edit ensemble_types.go to remove/update
	Members []Member `json:"members"`

	// After ensemble creation, how long should we reconcile
	// (in other words, how many seconds between checks?)
	// Defaults to 10 seconds
	// +kubebuilder:default=10
	// +default=10
	CheckSeconds int32 `json:"checkSeconds"`

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

	// Baseimage for the sidecar that will monitor the queue.
	// Ensure that the operating systems match!
	// +kubebuilder:default="ghcr.io/converged-computing/ensemble-operator-api:rockylinux9"
	// +default="ghcr.io/converged-computing/ensemble-operator-api:rockylinux9"
	// +optional
	SidecarBase string `json:"sidecarBase"`

	// Always pull the sidecar container (useful for development)
	// +optional
	SidecarPullAlways bool `json:"sidecarPullAlways"`

	// +kubebuilder:default="50051"
	// +default="50051"
	SidecarPort string `json:"sidecarPort"`

	// +kubebuilder:default=10
	// +default=10
	SidecarWorkers int32 `json:"sidecarWorkers"`

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
}

// EnsembleStatus defines the observed state of Ensemble
type EnsembleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// Validate ensures we have data that is needed, and sets defaults if needed
func (e *Ensemble) Validate() error {
	fmt.Println()

	// These are the allowed sidecars
	bases := set.New("ghcr.io/converged-computing/ensemble-operator-api:rockylinux9")

	// Global (entire cluster) settings
	fmt.Printf("🤓 Ensemble.members %d\n", len(e.Spec.Members))

	// If MaxSize is set, it must be greater than size
	if len(e.Spec.Members) < 1 {
		return fmt.Errorf("ensemble must have at least one member")
	}

	count := 0
	for i, member := range e.Spec.Members {

		fmt.Printf("   => Ensemble.member %d\n", i)

		// If we have a minicluster, all three sizes must be defined
		if !reflect.DeepEqual(member.MiniCluster, minicluster.MiniCluster{}) {
			fmt.Println("      Ensemble.member Type: minicluster")

			if member.SidecarBase == "" {
				member.SidecarBase = defaultSidecarbase
			}
			fmt.Printf("      Ensemble.member.SidecarBase: %s\n", member.SidecarBase)

			if member.MiniCluster.Spec.MaxSize <= 0 || member.MiniCluster.Spec.Size <= 0 {
				return fmt.Errorf("ensemble minicluster must have a size and maxsize of at least 1")
			}
			if member.MiniCluster.Spec.MinSize > member.MiniCluster.Spec.MaxSize {
				return fmt.Errorf("ensemble minicluster min size must be smaller than max size")
			}

			if member.MiniCluster.Spec.Size < member.MiniCluster.Spec.MinSize || member.MiniCluster.Spec.Size > member.MiniCluster.Spec.MaxSize {
				return fmt.Errorf("ensemble desired size must be between min and max size")
			}

			// Base container must be in valid set
			if !bases.Has(member.SidecarBase) {
				return fmt.Errorf("base image must be rocky linux or ubuntu: %s", bases)
			}
			count += 1
		}
	}
	// We shouldn't get here, but being pedantic
	if count == 0 {
		return fmt.Errorf("no members of the ensemble are valid")
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
