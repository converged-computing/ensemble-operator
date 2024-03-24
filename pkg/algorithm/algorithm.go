package algorithm

import (
	"encoding/json"
	"fmt"
	"log"

	api "github.com/converged-computing/ensemble-operator/api/v1alpha1"
)

// A lookup of registered algorithms by name
var (
	Algorithms   = map[string]AlgorithmInterface{}
	SubmitAction = "submit"
)

// An algorithm interface determines behavior for scaling and termination.
// Each algorithm can interact with the member gRPC client to:
//
//	get a current state or other metadata required to make a decision
//	directly influence the queue and submission (advanced)
type AlgorithmInterface interface {
	Name() string
	Description() string

	// Let's assume an algorithm can make a decision based on the gRPC payload
	MakeDecision(interface{}, []api.Job) (AlgorithmDecision, error)
	Validate(AlgorithmOptions) bool

	// Check that an algorithm is supported for a member type, and the member is valid
	Check(AlgorithmOptions, *api.Member) error
}

// An algorithm must return a decision for the operator to take
type AlgorithmDecision struct {

	// Scale up or down by size (e.g., negative value is size)
	Scale int32 `json:"scale"`

	// Terminate the member
	Terminate bool `json:"terminate"`

	// Send payload back to gRPC sidecar service
	Payload string `json:"payload"`

	// Action to ask the queue to take
	Action string `json:"action"`

	// Update determines if the spec was updated (warranting a patch)
	Updated bool `json:"updated"`
	Jobs    []api.Job
}

// AlgorithmOptions allow packaging named values of different types
// This is an alternative to using interfaces.
type AlgorithmOptions struct {
	BoolOpts map[string]bool
	StrOpts  map[string]string
	IntOpts  map[string]int32
	ListOpts map[string][]string
}

// ToJson serializes to json
func (e *AlgorithmDecision) ToJson() (string, error) {
	b, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), err
}

// List returns known backends
func List() map[string]AlgorithmInterface {
	return Algorithms
}

// Register a new backend by name
func Register(algorithm AlgorithmInterface) {
	if Algorithms == nil {
		Algorithms = make(map[string]AlgorithmInterface)
	}
	Algorithms[algorithm.Name()] = algorithm
}

// Get a backend by name
func Get(name string) (AlgorithmInterface, error) {
	for algoName, entry := range Algorithms {
		if algoName == name {
			return entry, nil
		}
	}
	return nil, fmt.Errorf("did not find algorithm named %s", name)
}

// GetOrFail ensures we can find the entry
func GetOrFail(name string) AlgorithmInterface {
	algorithm, err := Get(name)
	if err != nil {
		log.Fatalf("Failed to get algorithm: %v", err)
	}
	return algorithm
}
