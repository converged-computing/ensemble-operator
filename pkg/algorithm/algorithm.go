package algorithm

import (
	"encoding/json"
	"fmt"
	"log"

	api "github.com/converged-computing/ensemble-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// A lookup of registered algorithms by name
var (
	Algorithms = map[string]AlgorithmInterface{}

	// Operator Actions
	TerminateAction        = "terminate"
	CompleteAction         = "complete"
	ScaleAction            = "scale"
	ResetCounterAction     = "resetCounter"
	JobsMatrixUpdateAction = "updateJobsMatrix"

	// Queue actions
	SubmitAction  = "submit"
	JobInfoAction = "jobinfo"
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
	MakeDecision(*api.Ensemble, *api.Member, interface{}, []api.Job) (AlgorithmDecision, error)
	Validate(AlgorithmOptions) bool

	// Check that an algorithm is supported for a member type, and the member is valid
	Check(AlgorithmOptions, *api.Member) error
}

// An algorithm must return a decision for the operator to take
type AlgorithmDecision struct {

	// Scale up or down by size (e.g., negative value is size)
	Scale int32 `json:"scale"`

	// Send payload back to gRPC sidecar service
	Payload string `json:"payload"`

	// Action to ask the queue or operator to take
	Action string `json:"action"`

	// Jobs matrix
	Jobs []api.Job
}

// IsQueueRequest determines if the action should be sent to the queue
// Right now this is only the submit action
func (a *AlgorithmDecision) IsQueueRequest() bool {
	return a.Action == SubmitAction
}

// AlgorithmOptions allow packaging named values of different types
// This is an alternative to using interfaces.
type AlgorithmOptions map[string]intstr.IntOrString

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
