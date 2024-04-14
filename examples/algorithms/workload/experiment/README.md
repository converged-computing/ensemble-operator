# Workload Demand Experiment

You can read about the workload demand algorithm [here](https://github.com/converged-computing/ensemble-operator/blob/main/docs/algorithms.md#workoad-demand-of-consistent-sizes). Here we are doing a small experiment to test the following cases:

- static base case without ensemble (launching separate Miniclusters for each job) at max size
- autoscaling base case without ensemble (launching separate Miniclusters for each job) starting at min size, allowing scale up
- workload driven ensemble with autoscaler enabled and different submit approaches
  - random submit
  - ascending job size
  - descending job size

Note that there are two [autoscaling profiles](https://cloud.google.com/kubernetes-engine/docs/concepts/cluster-autoscaler#autoscaling_profiles) balanced (default) and optimize-utilization. I first tested balanced and found
that nodes hung around ~10 minutes after the queue was essentially empty, so I think the second one (that is noted to be more
aggressive) might be better. We are going to (as an extra bonus) keep track of the time the cluster takes to go back to the smallest
size when no work is running. I didn't see this was a parameter I could update.

## 1. Create the Cluster

These experiments are all run and controlled with python.

```bash
python run_experiments.py --help
python run_experiments.py --data-dir ./data/run0
```

TODO: do an experiment that looks at how the scale periods influence result!
