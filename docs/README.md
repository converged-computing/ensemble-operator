# Ensemble Operator

The ensemble operator is intended to run ensembles of workloads, and change them according to a user-specified [algorithm](algorithms.md).
You can learn more in the different markdown files below:

 - [User Guide](user-guide.md)
 - [Algorithms](algorithms.md)
 - [Design](design.md)
 - [Developer](developer.md)

Since an entity in an ensemble is typically more complex than a container, while we currently just support a Flux Operator MiniCluster, we could eventually
support a larger set of notable Kubernetes abstractions:

- [Job](https://kubernetes.io/docs/concepts/workloads/controllers/job/)
- [JobSet](https://github.com/kubernetes-sigs/jobset)
- [MiniCluster](https://github.com/flux-framework/flux-operator)
- [LeaderWorkersSet](https://github.com/kubernetes-sigs/lws)

These seem like a well-scoped set to start. For JobSet, LeaderWorkerSet, and MiniCluster, the corresponding operator for each is required to be installed to your cluster
to use them.
