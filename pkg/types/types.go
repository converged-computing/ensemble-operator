package types

// Types allows for serializing the queue and job status output from the gRPC sidecar

// MiniClusterStatus comes from flux, and looks like:
//
//	{
//	   "nodes": {
//	       "node_cores_free": 10,
//	       "node_cores_up": 10,
//	       "node_up_count": 1,
//	       "node_free_count": 1
//	   },
//	   "queue": {
//	       "new": 0,
//	       "depend": 0,
//	       "priority": 0,
//	       "sched": 0,
//	       "run": 0,
//	       "cleanup": 0,
//	       "inactive": 0
//	   }
//	}
type MiniClusterStatus struct {
	Nodes map[string]int32 `json:"nodes"`
	Queue map[string]int32 `json:"queue"`

	// Counts of things (e.g., number of checks we've done)
	Counts map[string]int32 `json:"counts"`
}
