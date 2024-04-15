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

This (larger) experiment has been moved to the [converged-computing/ensemble-experiments](https://github.com/converged-computing/ensemble-experiments) repository. This directory is used for testing components.


## 1. Create Cluster

We want to create a GKE cluster first. It won't have autoscaling enabled, etc.

```bash
GOOGLE_PROJECT=myproject
gcloud container clusters create test-cluster \
    --threads-per-core=1 \
    --placement-type=COMPACT \
    --num-nodes=6 \
    --region=us-central1-a \
    --project=${GOOGLE_PROJECT} \
    --machine-type=c2d-standard-8
```

Install the development operator and flux operator

```bash
make test-deploy-recreate
kubectl apply -f https://raw.githubusercontent.com/flux-framework/flux-operator/main/examples/dist/flux-operator.yaml
```

And apply, develop!

```bash
kubectl apply -f ensemble.yaml
```

When you are done, clean up.

```bash
gcloud container clusters delete test-cluster --region=us-central1-a
```