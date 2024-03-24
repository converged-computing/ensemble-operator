package demand

import (
	"fmt"

	api "github.com/converged-computing/ensemble-operator/api/v1alpha1"
	"github.com/converged-computing/ensemble-operator/internal/algorithm"
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
		return fmt.Errorf("algorithm %s is not supported for member type %s\n", AlgorithmName, memberType)
	}
	if !e.Validate(options) {
		return fmt.Errorf("algorithm options %s are not valid", AlgorithmName)
	}
	return nil
}

func (e WorkloadDemand) MakeDecision(
	payload interface{},
	member *api.Member,

) (algorithm.AlgorithmDecision, error) {

	fmt.Println(member.Jobs)
	decision := algorithm.AlgorithmDecision{}
	//file, err := os.ReadFile(yamlFile)
	//if err != nil {
	//		return &js, err
	//}

	return decision, nil
}

// Validate ensures that the sections provided are in the list we know
func (e WorkloadDemand) Validate(options algorithm.AlgorithmOptions) bool {
	return true
}

// Add the selection algorithm to be known to rainbow
func init() {
	a := WorkloadDemand{}
	algorithm.Register(a)
}
