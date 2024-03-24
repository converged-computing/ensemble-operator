package demand

// A SubmitRequest includes jobs and counts to submit.
// It will be serialized into the json payload to the gRPC sidecar
type SubmitRequest struct {
	Jobs []Job `json:"jobs,omitempty"`
}

type Job struct {
	Name    string `json:"name,omitempty"`
	Command string `json:"command,omitempty"`
	Count   int32  `json:"count,omitempty"`
	Nodes   int32  `json:"nodes,omitempty"`
}

func (r *SubmitRequest) AddJob(name, command string, count, nodes int32) {
	r.Jobs = append(r.Jobs, Job{Name: name, Command: command, Count: count, Nodes: nodes})
}
