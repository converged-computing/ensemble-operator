# Ensemble Operator

The ensemble operator is intended to run ensembles of workloads, and change them according to a user-specified [algorithm](algorithms.md).
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

Many details are TBA (still in my head) but you can read the following:

## Documentatino

 - [User Guide](user-guide.md)
 - [Algorithms](algorithms.md)
 - [Developer](developer.md)
