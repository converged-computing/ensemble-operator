package types

import (
	"math/rand"
)

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

	// List of 10 next jobs in queue
	NextJobs []int32 `json:"nextJobs"`

	// Waiting jobs lookup - node size by count
	Waiting map[int32]int32 `json:"waiting"`

	// Counts of things (e.g., number of checks we've done)
	Counts map[string]int32 `json:"counts"`
}

// GetLargestWaitingSize gets the largest size waiting
func (m *MiniClusterStatus) GetLargestWaitingSize() int32 {
	var maxNodes int32
	if len(m.Waiting) == 0 {
		return maxNodes
	}
	for nodes := range m.Waiting {
		if nodes > maxNodes {
			maxNodes = nodes
		}
	}
	return maxNodes
}

// GetSmallestWaitingSize gets the smallest node waiting size
func (m *MiniClusterStatus) GetSmallestWaitingSize() int32 {
	var minNodes int32
	if len(m.Waiting) == 0 {
		return minNodes
	}
	for nodes := range m.Waiting {
		if nodes < minNodes {
			minNodes = nodes
		}
	}
	return minNodes
}

// GetRandomWaitingSize returns a random size of a waiting job
func (m *MiniClusterStatus) GetRandomWaitingSize() int32 {
	selection := []int32{}
	for nodes, count := range m.Waiting {
		for i := 0; i < int(count); i++ {
			selection = append(selection, nodes)
		}
	}
	return selection[rand.Intn(len(selection))]
}
