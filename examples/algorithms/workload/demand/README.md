# Workload Demand

You can read about the workload demand algorithm [here](https://github.com/converged-computing/ensemble-operator/blob/main/docs/algorithms.md#workoad-demand-of-consistent-sizes).
For this example, we assume you have a cluster running (e.g., with kind) and have installed the Flux Operator and Ensemble Operator.

## Usage

Create the job. Note that this [ensemble.yaml](ensemble.yaml), for this algorithm type, requires you to submit jobs. It is retroactively going to adjust
the queue based on what you submit. We also are going to rely on the algorithm to determine when to terminate the minicluster, so we add a sleep command
to the end. Arguably this could be controlled (added) by the operator based on seeing the algorithm type, but this works for now.

```bash
kubectl apply -f ensemble.yaml
```

We can check both the gRPC sidecar and the operator to see if information is being received. Here is the
sidecar:

```bash
kubectl logs workload-demand-0-0-vfxxd -c api -f
```
```console
```

And here is the operator:

```bash
kubectl logs -n ensemble-operator-system ensemble-operator-controller-manager-5f874bb7d8-m68jb 
```


## TODO

- the base containers should have the grpc already built so we don't have to wait
- get a list of commands from the workload-demand decision endpoint, and send them back to the sidecar
  - this means we need to have a string payload that can be received
  - this also means the algorithms receiving endpoints need to be defined within the python service
- add to the decision logic to indicate a flag for when a spec (CRD) is updated so we do a patch