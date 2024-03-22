# Ensemble Operator

The ensemble operator is intended to run ensembles of workloads, and change them according to a user-specified algorithm.
Since an entity in an ensemble is typically more complex than a container, we allow creation of a few set of notable Kubernetes
abstractions:

- [Job](https://kubernetes.io/docs/concepts/workloads/controllers/job/)
- [JobSet](https://github.com/kubernetes-sigs/jobset)
- [MiniCluster](https://github.com/flux-framework/flux-operator)
- [LeaderWorkersSet](https://github.com/kubernetes-sigs/lws)

These seem like a well-scoped set to start. For JobSet, LeaderWorkerSet, and MiniCluster, the corresponding operator for each is required to be installed to your cluster
to use them. Thus, the default abstraction that will be created is Job, and for purposed of scaling out, typically an indexed job. Different kinds of operations to think about:

- change in size (e.g., a single Flux Operator Minicluster increasing or decreasing in size)
- scale (e.g., deploying more than one instance of a Job)

Details TBA, still in my head!

## Getting Started

Let's create a kind cluster first.

```bash
kind create cluster --config ./examples/kind-config.yaml
```

Ensure that the Flux Operator is installed.

```bash
kubectl apply -f https://raw.githubusercontent.com/flux-framework/flux-operator/main/examples/dist/flux-operator.yaml
```

And the ensemble operator

```bash
kubectl apply -f examples/dist/ensemble-operator.yaml
```

And then try the simple example to run lammps.

```bash
kubectl apply -f examples/tests/lammps/ensemble.yaml
```

This will create the MiniCluster, per the sizes you specified for it!

```bash
$ kubectl get pods
NAME                        READY   STATUS     RESTARTS   AGE
ensemble-sample-0-0-kc6qn   0/1     Init:0/1   0          3s
ensemble-sample-0-1-jjm4p   0/1     Init:0/1   0          3s
```

Next I will:

- better namespace the above
- decide if the sizes (min/max/desired) should be with the minicluster or member (where they are now)
- add the sidecar to the minicluster for the metrics api
- develop the algorithms for the flux metrics api to choose from
- make a cute logo

Then test it out! We will want different kinds of scaling, both inside and outside. I think I know what I'm going to do and just need to implement it.