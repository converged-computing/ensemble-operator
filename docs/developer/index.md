# Development

These pages (might?) include development and background for the process of creating
the Ensemble Operator. Right now they are just a few notes I took on the first day of working on it.

### Actions needed...

The operator should be notified under the following conditions:

- when to stop a MiniCluster (e.g., when is it done?)
- when to scale up
- when to scale down
- Note that the _cluster_ autoscaler has a concept of [expanders](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler/expander) that can be tied to request nodes for specific pools. The more advanced setup of this operator will also have a cluster autoscaler.

If you have any questions, please [let us know](https://github.com/converged-computing/ensemble-operator/issues)
