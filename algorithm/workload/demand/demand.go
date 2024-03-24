package demand

import (
	"encoding/json"
	"fmt"
	"math/rand"

	api "github.com/converged-computing/ensemble-operator/api/v1alpha1"
	"github.com/converged-computing/ensemble-operator/pkg/algorithm"
	"github.com/converged-computing/ensemble-operator/pkg/types"
	"k8s.io/utils/set"
)

// options for this algorithm include:
// 1. randomize for the initial jobs submission (defaults to true)
// 2. terminateChecks: number of checks (of an empty queue) to do as
//    an indicator for termination

const (
	AlgorithmName        = "workload-demand"
	AlgorithmDescription = "retroactively respond to workload needs"
)

// Member types that the algorithm supports
var (
	AlgorithmSupports      = set.New("minicluster")
	defaultRandomizeOption = true
	defaultTerminateChecks = 10
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
			req.AddJob(job.Name, job.Command, job.Count, job.Nodes)
			job.Count = 0
			updated = true
			updatedJobs[j] = job
		}
	}

	// Is there are request to not randomize?
	doRandomize := defaultRandomizeOption
	options := member.Algorithm.Options
	rOpt, ok := options["randomize"]
	if ok {
		if rOpt.StrVal == "no" {
			doRandomize = false
		}
	}

	// Randomly shuffle the jobs - this could be a parameter to the algorithm
	if doRandomize {
		if len(updatedJobs) > 0 {
			rand.Shuffle(len(updatedJobs), func(i, j int) { updatedJobs[i], updatedJobs[j] = updatedJobs[j], updatedJobs[i] })
		}
	}

	return &req, updatedJobs, updated
}

// MakeDecision for the workload-demand has the goal to:
// 1. Always submit all remaining jobs, in a random order
// 2. if we have exceeded
func (e WorkloadDemand) MakeDecision(
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
		terminate, err := e.terminateMember(member, payload)
		if err != nil {
			return decision, err
		}
		if terminate {
			decision.Action = algorithm.TerminateAction
		}
		return decision, nil

	} else {

		// Serialize the json into a payload
		response, err := json.Marshal(req)
		if err != nil {
			return decision, err
		}
		decision = algorithm.AlgorithmDecision{
			Updated: updated,
			Payload: string(response),
			Jobs:    updatedJobs,
		}
	}

	// If we have updates, ask the queue to submit them
	if updated {
		decision.Action = algorithm.SubmitAction
	}
	return decision, nil
}

// terminateMember uses the payload from the member to determine
// if we have reached termination criteria
func (e WorkloadDemand) terminateMember(member *api.Member, payload interface{}) (bool, error) {

	options := member.Algorithm.Options

	// Should we change the default number of checks?
	numberChecks := defaultTerminateChecks
	tOpt, ok := options["terminateChecks"]
	if ok {
		if tOpt.IntVal > 0 {
			numberChecks = tOpt.IntValue()
		}
	}

	// Parse the payload depending on the type
	if member.Type() == api.MiniclusterType {
		status := types.MiniClusterStatus{}
		err := json.Unmarshal([]byte(payload.(string)), &status)
		if err != nil {
			fmt.Printf("Error unmarshaling payload: %s\n", err)
			return false, err
		}
		fmt.Println(status)

		// Do we have an inactive count (subsequent times queue has not moved)?
		inactive, ok := status.Counts["inactive"]
		if !ok {
			fmt.Println("Cannot find inactive count")
			return false, nil
		}

		// Queue needs to be empty, nothing running, etc.
		activeJobs := status.Queue["new"] + status.Queue["priority"] + status.Queue["sched"] + status.Queue["run"] + status.Queue["cleanup"]

		// Conditions for termination:
		// 1. Inactive count exceeds our threshold
		// 2. Still no active jobs (above would be impossible if there were, but we double check)
		if inactive > int32(numberChecks) && activeJobs == 0 {
			fmt.Printf("Member %s is marked for termination\n", member.Type())
			return true, nil
		}

		// Here is where we terminate
		fmt.Printf("Member %s has active jobs or has not met threshold for for termination\n", member.Type())
		return false, nil
	}

	fmt.Printf("Warning: unknown member type %s\n", member.Type())
	return false, nil
}

// Validate ensures that the sections provided are in the list we know
func (e WorkloadDemand) Validate(options algorithm.AlgorithmOptions) bool {
	return true
}

func init() {
	a := WorkloadDemand{}
	algorithm.Register(a)
}
