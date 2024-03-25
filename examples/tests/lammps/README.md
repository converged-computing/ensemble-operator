# LAMMPS

You can read about the workload demand algorithm [here](https://github.com/converged-computing/ensemble-operator/blob/main/docs/algorithms.md#workoad-demand-of-consistent-sizes).
For this example, you should create the provided kind cluster (4 nodes):

```bash
kind create cluster --config ./examples/dist/ensemble-operator.yaml
```

And install the operators:

```bash
kubectl apply -f https://raw.githubusercontent.com/flux-framework/flux-operator/main/examples/dist/flux-operator.yaml
kubectl apply -f https://raw.githubusercontent.com/converged-computing/ensemble-operator/main/examples/dist/ensemble-operator.yaml
```

Here is what is going to happen:

1. We define our jobs matrix in the [ensemble.yaml](ensemble.yaml). The jobs matrix is consistent across algorithms.
2. For any algorithm, we first get a cluster status `StatusRequest`. We don't use it here aside from establishing communication. In the future if we return more metadata about the cluster it can be used to inform decision making.
3. We then detect that the job matrix has outstanding jobs, and make an `ActionRequest` to "submit" the jobs.
4. The jobs are submit on the MiniCluster, and the matrix is emptied.
5. We proceed to monitor, scaling when conditions are met, downsizing when jobs are finishing, and terminating after that.

## Usage

Create the job. Note that this [ensemble.yaml](ensemble.yaml), for this algorithm type, requires you to submit jobs. It is retroactively going to adjust
the queue based on what you submit. We also are going to rely on the algorithm to determine when to terminate the minicluster, so we add a sleep command
to the end. Arguably this could be controlled (added) by the operator based on seeing the algorithm type, but this works for now.

```bash
kubectl apply -f ensemble.yaml
```

We can check both the gRPC sidecar and the operator to see if information is being received. Here is the
sidecar (after setup steps):

```bash
# This is for the sidecar
kubectl logs workload-demand-0-0-vfxxd -c api -f

# This is for the ensemble operator
$ kubectl logs workload-demand-0-0-d2trb -c api -f
```

You'll see:

1. The jobs being submit at the onset
2. Each request from the Ensemble Operator to get queue status
3. When we reach the threshold of the queue not moving after 2 checks, we scale up
4. The jobs complete and the ensemble terminates.


