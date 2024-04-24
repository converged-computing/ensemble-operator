# Workload Demand Autoscaling Experiment

You can read about the workload demand algorithm [here](https://github.com/converged-computing/ensemble-operator/blob/main/docs/algorithms.md#workoad-demand-of-consistent-sizes). Here we are testing this algorithm with autoscaling.

This (larger) experiment has been moved to the [converged-computing/ensemble-experiments](https://github.com/converged-computing/ensemble-experiments) repository. This directory is used for testing components.

## Create Cluster

We want to create a GKE cluster first. It won't have autoscaling enabled, etc.

```bash
GOOGLE_PROJECT=myproject
gcloud container clusters create test-cluster \
    --enable-autoscaling \
    --threads-per-core=1 \
    --placement-type=COMPACT \
    --autoscaling-profile=optimize-utilization \
    --region=us-central1-a \
    --num-nodes 1 \
    --total-min-nodes 1 \
    --total-max-nodes 18 \
    --project=${GOOGLE_PROJECT} \
    --machine-type=c2d-standard-8
```

Install the development operator and flux operator

```bash
make test-deploy-recreate
kubectl apply -f https://raw.githubusercontent.com/flux-framework/flux-operator/main/examples/dist/flux-operator.yaml
```

## Run the Ensemble

And apply, develop!

```bash
kubectl apply -f ensemble.yaml
```

When you are done, clean up.

```bash
gcloud container clusters delete test-cluster --region=us-central1-a
```
