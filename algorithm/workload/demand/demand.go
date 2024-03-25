package demand

import (
	"encoding/json"
	"fmt"
	"math/rand"

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
	AlgorithmSupports      = set.New("minicluster")
	defaultRandomizeOption = true
	defaultTerminateChecks = 10
	defaultScaleChecks     = 5

	scaleDownStrategyExcess    = "excess"
	scaleUpStrategyNextJob     = "nextJob"
	scaleUpStrategySmallestJob = "smallestJob"
	scaleUpStrategyLargestJob  = "largestJob"
	scaleUpStrategyRandomJob   = "randomJob"
	defaultScaleUpStrategy     = scaleUpStrategyNextJob
	defaultScaleDownStrategy   = scaleDownStrategyExcess
)

// Workload demand algorithm takes two options
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

// getUpdatedJobs returns jobs that haven't been run yet
// We randomly sort them.
func (e WorkloadDemand) getUpdatedJobs(
	member *api.Member,
	jobs []api.Job,
) (*SubmitRequest, []api.Job, bool) {

	// Assume not updated at first
	updated := false
	req := SubmitRequest{Jobs: []Job{}}
	updatedJobs := make([]api.Job, len(jobs))
	for j, job := range jobs {
		if job.Count > 0 {
			req.AddJob(job)
			job.Count = 0
			updated = true
			updatedJobs[j] = job
		}
	}

	// Do we want to randomize by group or job?
	doRandomize := member.StringToBooleanOption("randomize", defaultRandomizeOption)
	req.Randomize = doRandomize

	// Randomly shuffle the job groups - random shufflig of ALL jobs happens at the queue level
	// For efficient transfer of matrix
	if doRandomize {
		if len(updatedJobs) > 0 {
			rand.Shuffle(len(updatedJobs), func(i, j int) { updatedJobs[i], updatedJobs[j] = updatedJobs[j], updatedJobs[i] })
		}
	}

	return &req, updatedJobs, updated
}

// MakeDecision for the workload-demand determines if we want to submit
// jobs (the inital request will determine this is needed), terminate
// or complete a member (based on an inactive state over some number of
// iterations) or scale the cluster up or down.
func (e WorkloadDemand) MakeDecision(
	ensemble *api.Ensemble,
	member *api.Member,
	payload interface{},
	jobs []api.Job,

) (algorithm.AlgorithmDecision, error) {

	// This will determine if we need to patch the crd
	decision := algorithm.AlgorithmDecision{}

	// Get the submit request, and updated jobs
	req, updatedJobs, updated := e.getUpdatedJobs(member, jobs)

	// Only look to do another action (scale, terminate, etc) if we don't have updated jobs
	if !updated {

		// Check for termination or completion first
		terminate, done, err := e.terminateMember(member, payload)
		if err != nil {
			return decision, err
		}

		// This means we are done, stop checking but do not terminate
		if done {
			decision.Action = algorithm.CompleteAction
		} else if terminate {
			decision.Action = algorithm.TerminateAction
		}

		// If we have a terminate or complete action, return it
		if decision.Action != "" {
			return decision, nil
		}

		// Otherwise, continue and check for scaling
		scale, err := e.scaleMember(member, payload)
		if err != nil {
			return decision, err
		}

		// Nonzero indicates a decision to scale up or down
		if scale != 0 {
			decision.Action = algorithm.ScaleAction
			decision.Scale = scale
		}
		return decision, err

	} else {

		// UpdatedJobs warrant the submit action
		// Serialize the json into a payload
		response, err := json.Marshal(req)
		if err != nil {
			return decision, err
		}
		decision = algorithm.AlgorithmDecision{

			// This will be handed to the cluster as the SubmitAction
			Action:  algorithm.JobsMatrixUpdateAction,
			Payload: string(response),
			Jobs:    updatedJobs,
		}
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
