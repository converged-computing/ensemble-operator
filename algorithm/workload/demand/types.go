package demand

import api "github.com/converged-computing/ensemble-operator/api/v1alpha1"

// A SubmitRequest includes jobs and counts to submit.
// It will be serialized into the json payload to the gRPC sidecar
type SubmitRequest struct {
	Jobs  []Job  `json:"jobs,omitempty"`
	Order string `json:"order"`
}

// Flatten a job out from the matrix structure
type Job struct {
	Name    string `json:"name,omitempty"`
	Command string `json:"command,omitempty"`
	Nodes   int32  `json:"nodes,omitempty"`

	// Tasks defaults to one, and node count cannot
	// be greater than task count
	Tasks    int32  `json:"tasks,omitempty"`
	Workdir  string `json:"workdir,omitempty"`
	Duration int32  `json:"duration,omitempty"`
	Count    int32  `json:"count,omitempty"`
}

// AddJob (unwrapping from matrix format) into a submit request
// TODO look into limit size of GRPC and see how many jobs we can fit
func (r *SubmitRequest) AddJob(job api.Job) {
	newJob := Job{
		Name:     job.Name,
		Command:  job.Command,
		Nodes:    job.Nodes,
		Workdir:  job.Workdir,
		Duration: job.Duration,
		Tasks:    job.Tasks,
		Count:    job.Count,
	}
	r.Jobs = append(r.Jobs, newJob)
}
