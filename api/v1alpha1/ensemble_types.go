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
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/intstr"

	minicluster "github.com/flux-framework/flux-operator/api/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/set"
)

var (
	defaultSidecarbase   = "ghcr.io/converged-computing/ensemble-operator-api:rockylinux9"
	defaultAlgorithmName = "workload-demand"

	MiniclusterType = "minicluster"
	UnknownType     = "unknown"
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
	//+optional
	Algorithm Algorithm `json:"algorithm"`
}

// A member of the ensemble that will run for some number of times,
// optionally with a maximum or minumum
type Member struct {

	// MiniCluster is of a type MiniCluster, the base unit of an ensemble.
	// We do this because we install a flux metrics API within each MiniCluster to manage it
	// TODO where should the user define the size? Here or with the member?
	// +optional
	MiniCluster minicluster.MiniCluster `json:"minicluster,omitempty"`

	// Definition and customization of the sidecar
	//+optional
	Sidecar Sidecar `json:"sidecar,omitempty"`

	// A member is required to define one or more jobs
	// These are passed into status for further updating
	// Jobs
	Jobs []Job `json:"jobs"`

	// Member specific algorithm to use
	// If not defined, defaults to workload-demand
	//+optional
	Algorithm Algorithm `json:"algorithm"`
}

type Sidecar struct {

	// Baseimage for the sidecar that will monitor the queue.
	// Ensure that the operating systems match!
	// +kubebuilder:default="ghcr.io/converged-computing/ensemble-operator-api:rockylinux9"
	// +default="ghcr.io/converged-computing/ensemble-operator-api:rockylinux9"
	// +optional
	Image string `json:"image"`

	// Always pull the sidecar container (useful for development)
	// +optional
	PullAlways bool `json:"pullAlways"`

	// +kubebuilder:default="50051"
	// +default="50051"
	Port string `json:"port"`

	// +kubebuilder:default=10
	// +default=10
	Workers int32 `json:"workers"`
}

type Algorithm struct {

	// +kubebuilder:default="workload-demand"
	// +default="workload-demand"
	//+optional
	Name string `json:"name"`

	// Options for the algorithm
	//+optional
	Options map[string]intstr.IntOrString `json:"options"`
}

// Job defines a unit of work for the ensemble to munch on. Munch munch munch.
type Job struct {

	// Name to identify the job group
	Name string `json:"name"`

	// Command given to flux
	Command string `json:"command"`

	// Number of jobs to run
	// This can be set to 0 depending on the algorithm
	// E.g., some algorithms decide on the number to submit
	// +kubebuilder:default=1
	// +default=1
	//+optional
	Count int32 `json:"count"`

	// Number of nodes for job
	// +kubebuilder:default=1
	// +default=1
	//+optional
	Nodes int32 `json:"nodes"`

	// Number of tasks for the job, defaults to 1
	// Node count cannot be greater than task count
	//+optional
	Tasks int32 `json:"tasks"`

	// Working directory for job group (if unset, uses container default)
	//+optional
	Workdir string `json:"workdir"`

	// Duration of jobs
	// Value of 0 (default) indicates no limit
	//+optional
	Duration int32 `json:"duration"`
}

// EnsembleStatus defines the observed state of Ensemble
type EnsembleStatus struct {

	// These lookup values must be string, so we convert from int32 to it
	// Jobs matrix lookup (what we can mutate)
	Jobs map[string][]Job `json:"jobs"`
}

// Helper function get member type
func (m *Member) Type() string {
	if !reflect.DeepEqual(m.MiniCluster, minicluster.MiniCluster{}) {
		return MiniclusterType
	}
	return UnknownType
}

// getStringOption returns a string option from the member
func (m *Member) StringToBooleanOption(
	key string,
	defaultValue bool,
) bool {

	value := defaultValue
	options := m.Algorithm.Options
	rOpt, ok := options[key]
	if ok {
		if rOpt.StrVal == "no" {
			value = false
		} else if rOpt.StrVal == "yes" {
			value = true
		}
	}
	return value
}

// getStringOption returns a string option from the member
func (m *Member) GetStringOption(
	key string,
	defaultValue string,
) string {

	value := defaultValue
	options := m.Algorithm.Options
	rOpt, ok := options[key]
	if ok {
		value = rOpt.StrVal
	}
	return value
}

// getPositiveIntegerOption returns an option > 0
func (m *Member) GetPositiveIntegerOption(
	key string,
	defaultValue int,
) int {

	options := m.Algorithm.Options
	value := defaultValue
	tOpt, ok := options[key]
	if ok {
		if tOpt.IntVal > 0 {
			value = tOpt.IntValue()
		}
	}
	return value
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

func (e *Ensemble) getDefaultAlgorithm() Algorithm {
	defaultAlgorithm := e.Spec.Algorithm

	// No we don't, it's empty
	if reflect.DeepEqual(defaultAlgorithm, Algorithm{}) {
		defaultAlgorithm = Algorithm{Name: defaultAlgorithmName}
	}
	return defaultAlgorithm
}

func (e *Ensemble) RequeueAfter() time.Duration {
	return time.Duration(time.Duration(e.Spec.CheckSeconds) * time.Second)
}

// Validate ensures we have data that is needed, and sets defaults if needed
func (e *Ensemble) Validate() error {
	fmt.Println()

	// These are the allowed sidecars
	bases := set.New(
		"ghcr.io/converged-computing/ensemble-operator-api:rockylinux9-test",
		"ghcr.io/converged-computing/ensemble-operator-api:rockylinux9",
		"ghcr.io/converged-computing/ensemble-operator-api:rockylinux8",
		"ghcr.io/converged-computing/ensemble-operator-api:ubuntu-focal",
		"ghcr.io/converged-computing/ensemble-operator-api:ubuntu-jammy",
	)

	// Global (entire cluster) settings
	fmt.Printf("🤓 Ensemble.members %d\n", len(e.Spec.Members))

	// Do we have a default algorithm set?
	defaultAlgorithm := e.getDefaultAlgorithm()

	// If MaxSize is set, it must be greater than size
	if len(e.Spec.Members) < 1 {
		return fmt.Errorf("ensemble must have at least one member")
	}

	count := 0
	for i, member := range e.Spec.Members {

		fmt.Printf("   => Ensemble.member %d\n", i)

		// If we don't have an algorithm set, use the default
		if reflect.DeepEqual(defaultAlgorithm, Algorithm{}) {
			member.Algorithm = defaultAlgorithm
		}
		fmt.Printf("      Ensemble.member.Algorithm: %s\n", member.Algorithm.Name)

		// The member must have at least one job definition
		if len(member.Jobs) == 0 {
			return fmt.Errorf("ensemble member in index %d must have at least one job definition", i)
		}

		// Validate jobs matrix
		for _, job := range member.Jobs {
			if job.Count <= 0 {
				job.Count = 1
			}
			if job.Tasks != 0 && job.Nodes > job.Tasks {
				return fmt.Errorf("issue with job matrix for ensemble member in index %d: node count cannot be greater than task count", i)
			}
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

			// MiniCluster specific validation for jobs - we cannot go over max size
			for _, job := range member.Jobs {
				if job.Nodes > member.MiniCluster.Spec.MaxSize {
					return fmt.Errorf("job node requirement cannot exceed max size of %d", member.MiniCluster.Spec.MaxSize)
				}
			}

			fmt.Println("      Ensemble.member Type: minicluster")
			if member.Sidecar.Image == "" {
				member.Sidecar.Image = defaultSidecarbase
			}
			fmt.Printf("      Ensemble.member.Sidecar.Image: %s\n", member.Sidecar.Image)
			fmt.Printf("      Ensemble.member.Sidecar.Port: %s\n", member.Sidecar.Port)
			fmt.Printf("      Ensemble.member.Sidecar.PullAlways: %v\n", member.Sidecar.PullAlways)

			if member.MiniCluster.Spec.MaxSize <= 0 || member.MiniCluster.Spec.Size <= 0 {
				return fmt.Errorf("ensemble minicluster must have a size and maxsize of at least 1")
			}
			if member.MiniCluster.Spec.MinSize > member.MiniCluster.Spec.MaxSize {
				return fmt.Errorf("ensemble minicluster min size must be smaller than max size")
			}

			if member.MiniCluster.Spec.Size < member.MiniCluster.Spec.MinSize || member.MiniCluster.Spec.Size > member.MiniCluster.Spec.MaxSize {
				return fmt.Errorf("ensemble desired size must be between min and max size")
			}

			// If the sidecar has a digest it's OK, but remove for check
			sidecar := member.Sidecar.Image
			if strings.Contains(sidecar, "@") {
				parts := strings.Split(sidecar, "@")
				sidecar = parts[0]
			}

			// Base container must be in valid set
			if !bases.Has(sidecar) {
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
