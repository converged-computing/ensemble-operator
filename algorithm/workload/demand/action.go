package demand

import (
	"encoding/json"
	"fmt"

	api "github.com/converged-computing/ensemble-operator/api/v1alpha1"
	"github.com/converged-computing/ensemble-operator/pkg/types"
)

// scaleMember uses the payload from the member to determine
// if we have reached a scaling criteria state
func (e WorkloadDemand) scaleMember(
	member *api.Member,
	payload interface{},
) (int32, error) {

	// Parse the payload depending on the type
	if member.Type() == api.MiniclusterType {

		status := types.MiniClusterStatus{}
		err := json.Unmarshal([]byte(payload.(string)), &status)
		if err != nil {
			fmt.Printf("      Error unmarshaling payload: %s\n", err)
			return 0, err
		}

		// This is the scaling strategy to use for going up
		scaleUpStrategy := member.GetStringOption("scaleUpStrategy", defaultScaleUpStrategy)

		// First check if we have enough checks done to warrant doing a scale, period
		scaleChecks := member.GetPositiveIntegerOption("scaleChecks", defaultScaleChecks)

		// scaling up is dependent on number of waiting periods at or == an original value
		waitingPeriod, ok := status.Counts["waiting_periods"]
		var scaleUpBy int32
		if ok && len(status.Waiting) > 0 {

			fmt.Println("🍔️ Scaling event")
			fmt.Printf("     => Scale up strategy: %s\n", scaleUpStrategy)

			// If the number of waiting periods exceeds our threshold, this is a scaling event
			if waitingPeriod >= int32(scaleChecks) {

				// Choose the smallest job
				if scaleUpStrategy == scaleUpStrategySmallestJob {

					// This scale payload has a lookup of waiting nodes -> counts
					// The smallest key is the smallest size waiting, etc.
					scaleUpBy = status.GetSmallestWaitingSize()
					fmt.Printf("        Scaling up by: %d\n", scaleUpBy)
					return scaleUpBy, nil

				} else if scaleUpStrategy == scaleUpStrategyLargestJob {
					scaleUpBy = status.GetLargestWaitingSize()
					fmt.Printf("        Scaling up by: %d\n", scaleUpBy)
					return scaleUpBy, nil
				} else if scaleUpStrategy == scaleUpStrategyRandomJob {
					scaleUpBy = status.GetRandomWaitingSize()
					fmt.Printf("        Scaling up by: %d\n", scaleUpBy)
					return scaleUpBy, nil
				} else {

					// Catch all for else (defautl) and this includes a scale strategy
					// that isn't defined (user error)
					// We can only do this if we have next jobs!
					if len(status.NextJobs) > 0 {
						scaleUpBy := status.NextJobs[0]
						fmt.Printf("        Scaling up by: %d\n", scaleUpBy)
						return scaleUpBy, nil
					}
				}
			}
		} else {
			fmt.Printf("        Waiting jobs count %d does not warrant scaling up\n", len(status.Waiting))
		}
	}
	// TODO implement scale down
	//	"smallestJob" "nextJob" "randomJob"
	return 0, nil
}

// terminateMember uses the payload from the member to determine
// if we have reached termination criteria, or completion criteria
func (e WorkloadDemand) terminateMember(
	member *api.Member,
	payload interface{},
) (bool, bool, error) {

	// Do we want to skip termination? We use this to indicate to the operator
	// to stop checking when it finds we would have terminated.
	skipTerminate := member.StringToBooleanOption("disableTermination", false)

	// Get the number of termination checks to do
	numberChecks := member.GetPositiveIntegerOption("terminateChecks", defaultTerminateChecks)

	// Parse the payload depending on the type
	if member.Type() == api.MiniclusterType {

		status := types.MiniClusterStatus{}
		err := json.Unmarshal([]byte(payload.(string)), &status)
		if err != nil {
			fmt.Printf("Error unmarshaling payload: %s\n", err)
			return false, false, err
		}
		fmt.Println(status)

		// Do we have an inactive count (subsequent times queue has not moved)?
		inactive, ok := status.Counts["inactive"]
		if !ok {
			fmt.Println("Cannot find inactive count")
			return false, false, nil
		}

		// Queue needs to be empty, nothing running, etc.
		activeJobs := status.Queue["new"] + status.Queue["priority"] + status.Queue["sched"] + status.Queue["run"] + status.Queue["cleanup"]

		// Conditions for termination:
		// 1. Inactive count exceeds our threshold
		// 2. Still no active jobs (above would be impossible if there were, but we double check)
		if inactive > int32(numberChecks) && activeJobs == 0 {
			if skipTerminate {
				fmt.Printf("Member %s is completed\n", member.Type())
				return false, true, nil
			}
			fmt.Printf("Member %s is marked for termination\n", member.Type())
			return true, false, nil
		}

		// Here is where we terminate
		fmt.Printf("Member %s has active jobs or has not met threshold for for termination\n", member.Type())
		return false, false, nil
	}

	fmt.Printf("Warning: unknown member type %s\n", member.Type())
	return false, false, nil
}
