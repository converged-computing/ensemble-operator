package demand

import (
	"encoding/json"
	"fmt"

	api "github.com/converged-computing/ensemble-operator/api/v1alpha1"
	"github.com/converged-computing/ensemble-operator/pkg/algorithm"
	"k8s.io/utils/set"
)

const (
	AlgorithmName        = "workload-demand"
	AlgorithmDescription = "retroactively respond to workload needs"
)

// Member types that the algorithm supports
var (
	AlgorithmSupports = set.New("minicluster")
)

type WorkloadDemand struct{}

func (e WorkloadDemand) Name() string {
	return AlgorithmName
}

func (e WorkloadDemand) Description() string {
	return AlgorithmDescription
}

// Check that an algorithm is supported for a member type and has valid options
func (e WorkloadDemand) Check(
	options algorithm.AlgorithmOptions,
	member *api.Member,
) error {
	memberType := member.Type()
	if !AlgorithmSupports.Has(memberType) {
		return fmt.Errorf("algorithm %s is not supported for member type %s", AlgorithmName, memberType)
	}
	if !e.Validate(options) {
		return fmt.Errorf("algorithm options %s are not valid", AlgorithmName)
	}
	return nil
}

func (e WorkloadDemand) MakeDecision(
	payload interface{},
	jobs []api.Job,

) (algorithm.AlgorithmDecision, error) {

	// This will determine if we need to patch the crd
	decision := algorithm.AlgorithmDecision{}
	updated := false

	// For the algorithm here, we always submit all jobs that still have counts.
	// We don't consider the queue status (payload) here because it does not matter
	// This is a lookup
	req := SubmitRequest{Jobs: []Job{}}
	updatedJobs := make([]api.Job, len(jobs))
	for j, job := range jobs {
		if job.Count > 0 {
			req.AddJob(job.Name, job.Command, job.Count, job.Nodes)
			job.Count = 0
			updated = true
			updatedJobs[j] = job
		}
	}

	// Serialize the json into a payload
	response, err := json.Marshal(&req)
	if err != nil {
		return decision, err
	}
	decision = algorithm.AlgorithmDecision{Updated: updated, Payload: string(response), Jobs: updatedJobs}

	// If we have updates, ask the queue to submit them
	if updated {
		decision.Action = algorithm.SubmitAction
	}
	return decision, nil
}

// Validate ensures that the sections provided are in the list we know
func (e WorkloadDemand) Validate(options algorithm.AlgorithmOptions) bool {
	return true
}

func init() {
	a := WorkloadDemand{}
	algorithm.Register(a)
}
