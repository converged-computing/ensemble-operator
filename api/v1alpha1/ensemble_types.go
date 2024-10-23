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
)

var (
	defaultSidecarbase = "ghcr.io/converged-computing/ensemble-python:latest"
	MiniclusterType    = "minicluster"
	UnknownType        = "unknown"
)

// EnsembleSpec defines the desired state of Ensemble
type EnsembleSpec struct {
	Members []Member `json:"members"`

	// Definition and customization of the sidecar
	//+optional
	Sidecar Sidecar `json:"sidecar,omitempty"`
}

// A member of the ensemble that will run for some number of times,
// optionally with a maximum or minumum
type Member struct {

	// MiniCluster is of a type MiniCluster, the base unit of an ensemble.
	// We do this because we install a flux metrics API within each MiniCluster to manage it
	// TODO where should the user define the size? Here or with the member?
	// +optional
	MiniCluster minicluster.MiniCluster `json:"minicluster,omitempty"`

	// Branch
	// Instead of pip, install a specific branch of ensemble python
	// +optional
	Branch string `json:"branch"`

	// Ensemble yaml (configuration file)
	Ensemble string `json:"ensemble"`
}

type Sidecar struct {

	// Baseimage for the sidecar that will monitor the queue.
	// Ensure that the operating systems match!
	// +kubebuilder:default="ghcr.io/converged-computing/ensemble-operator-api:rockylinux9"
	// +default="ghcr.io/converged-computing/ensemble-operator-api:rockylinux9"
	// +optional
	Image string `json:"image"`

	// Sidecar image pull policy
	// +optional
	ImagePullPolicy string `json:"imagePullPolicy"`

	// +kubebuilder:default="50051"
	// +default="50051"
	Port string `json:"port"`

	// +kubebuilder:default=10
	// +default=10
	Workers int32 `json:"workers"`
}

// EnsembleStatus defines the observed state of Ensemble
type EnsembleStatus struct{}

// Helper function get member type
func (m *Member) Type() string {
	if !reflect.DeepEqual(m.MiniCluster, minicluster.MiniCluster{}) {
		return MiniclusterType
	}
	return UnknownType
}

// Size is a common function to return a member size
// This should only be used on init, as the size is then stored in status
// As long as the MiniCluster is not created, the actual spec size won't
// be used again.
func (m *Member) Size() int32 {
	if !reflect.DeepEqual(m.MiniCluster, minicluster.MiniCluster{}) {
		return m.MiniCluster.Spec.Size
	}
	return 0
}

func (e *Ensemble) ServiceName() string {
	return fmt.Sprintf("%s-grpc", e.Name)
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

	// Validate the sidecar deployment
	if e.Spec.Sidecar.Image == "" {
		e.Spec.Sidecar.Image = defaultSidecarbase
	}

	fmt.Printf("      Ensemble.Sidecar.Image: %s\n", e.Spec.Sidecar.Image)
	fmt.Printf("      Ensemble.Sidecar.Port: %s\n", e.Spec.Sidecar.Port)
	if e.Spec.Sidecar.ImagePullPolicy != "" {
		fmt.Printf("      Ensemble.Sidecar.ImagePullPolicy: %v\n", e.Spec.Sidecar.ImagePullPolicy)
	}

	// TODO stopped here - make interactive cluster with grpc running, shell in, and test
	// client.
	count := 0
	for i, member := range e.Spec.Members {

		fmt.Printf("   => Ensemble.member %d\n", i)

		// Every member needs an ensemble, the yaml file, no exceptions.
		if member.Ensemble == "" {
			return fmt.Errorf("member in index %d is missing the ensemble (yaml) spec string", i)
		}

		// If we have a minicluster, all three sizes must be defined
		if !reflect.DeepEqual(member.MiniCluster, minicluster.MiniCluster{}) {

			// If they don't set it, they get a very small size :)
			if member.MiniCluster.Spec.Size <= 0 {
				member.MiniCluster.Spec.Size = 1
			}
			// If MaxSize is not set, assume it's the size
			// The Flux Operator does other validation checks, but we need this here!
			if member.MiniCluster.Spec.MaxSize == 0 {
				member.MiniCluster.Spec.MaxSize = member.MiniCluster.Spec.Size
			}

			// If no default image, consider an error
			if member.MiniCluster.Spec.Containers[0].Image == "" {
				return fmt.Errorf("ensemble minicluster must have an image")
			}
			fmt.Println("      Ensemble.member Type: minicluster")

			if member.MiniCluster.Spec.MaxSize <= 0 || member.MiniCluster.Spec.Size <= 0 {
				return fmt.Errorf("ensemble minicluster must have a size and maxsize of at least 1")
			}
			if member.MiniCluster.Spec.MinSize > member.MiniCluster.Spec.MaxSize {
				return fmt.Errorf("ensemble minicluster min size must be smaller than max size")
			}

			if member.MiniCluster.Spec.Size < member.MiniCluster.Spec.MinSize || member.MiniCluster.Spec.Size > member.MiniCluster.Spec.MaxSize {
				return fmt.Errorf("ensemble desired size must be between min and max size")
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
