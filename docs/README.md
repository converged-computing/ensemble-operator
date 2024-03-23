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

### 1. Create Cluster

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

### 2. Run LAMMPS

And then try the simple example to run lammps.

```bash
kubectl apply -f examples/tests/lammps/ensemble.yaml
```

This will create the MiniCluster, per the sizes you specified for it!

```bash
$ kubectl get pods
```
```console
NAME                        READY   STATUS     RESTARTS   AGE
ensemble-sample-0-0-kc6qn   0/1     Init:0/1   0          3s
ensemble-sample-0-1-jjm4p   0/1     Init:0/1   0          3s
```

You'll first see init containers (above) that are preparing the flux install. When the containers are running,
you'll then see two containers:

```console
NAME                        READY   STATUS    RESTARTS   AGE
ensemble-sample-0-0-zhg47   2/2     Running   0          44s
ensemble-sample-0-1-6dpgm   2/2     Running   0          44s
```

### 3. Check GRPC Service Endpoint

We have two things that are working together:

- The *GRPC service endpoint* is being served by a sidecar container alongside the MiniCluster
- The *GRPC client* is created by the Ensemble operator by way of looking up the pod ip address

TLDR: the operator can look at the status of the ensemble queue because a grpc service pod is running alongside the MiniCluster, and providing an endpoint that has direct access to the queue there! We can then implement and choose some algorithm to decide how to scale or terminate the ensemble.
Let's now check that this started correctly - "api" is the name of the container running the sidecar GRPC service:

```bash
kubectl logs ensemble-sample-0-0-zhg47 -c api -f
```
```console
[notice] A new release of pip is available: 23.2.1 -> 24.0
[notice] To update, run: pip3 install --upgrade pip
🥞️ Starting ensemble endpoint at :50051
```

We can also check the GRPC endpoint from the operator - depending on when you check, you'll see the payload delivered!

```bash
kubectl logs -n ensemble-operator-system ensemble-operator-controller-manager-5f874bb7d8-2sbcp -f
```
```console
2024/03/23 01:43:55 🥞️ starting client (10.244.3.23:50051)...
&{10.244.3.23:50051 0xc000077800 0xc0006ae2f0}
payload:"{\"nodes\": {\"node_cores_free\": 18, \"node_cores_up\": 20, \"node_up_count\": 2, \"node_free_count\": 2}, \"queue\": {\"RUN\": 1, \"new\": 0, \"depend\": 0, \"priority\": 0, \"sched\": 0, \"run\": 0, \"cleanup\": 0, \"inactive\": 0}}"  status:SUCCESS
SUCCESS
{"nodes": {"node_cores_free": 18, "node_cores_up": 20, "node_up_count": 2, "node_free_count": 2}, "queue": {"RUN": 1, "new": 0, "depend": 0, "priority": 0, "sched": 0, "run": 0, "cleanup": 0, "inactive": 0}}
2024-03-23T01:43:55Z    INFO    🥞️ Ensemble is Ready!   {"controller": "ensemble", "controllerGroup": "ensemble.flux-framework.org", "controllerKind": "Ensemble", "Ensemble": {"name":"ensemble-sample","namespace":"default"}, "namespace": "default", "name": "ensemble-sample", "reconcileID": "8ca7973f-17f3-478c-a15b-7d125ca646cd"}
```

That output is not parsed (so not pretty yet) but it will be! An Algorithm interface (TBA) will accept that state, and then decide on an action to take. Keep reading the Developer sections below for the high level actions we might do.
And you can see the pings in the client to. They will be at the frequency you specified for your Ensemble CheckSeconds (defaults to 10)

```bash
kubectl logs ensemble-sample-0-0-dwr2h -c api -f
```
```console
[notice] A new release of pip is available: 23.2.1 -> 24.0
[notice] To update, run: pip3 install --upgrade pip
🥞️ Starting ensemble endpoint at :50051

<grpc._server._Context object at 0x7f699aaef690>
{
    "nodes": {
        "node_cores_free": 10,
        "node_cores_up": 10,
        "node_up_count": 1,
        "node_free_count": 1
    },
    "queue": {
        "new": 0,
        "depend": 0,
        "priority": 0,
        "sched": 0,
        "run": 0,
        "cleanup": 0,
        "inactive": 0
    }
}
```

In practice this means we are putting more burden on our operator to keep reconciling when it might finish and stop. But also for this use case of running HPC jobs, I think it's more likely to have a smaller number of ensembles running vs. hundreds of thousands of them. Anyway, scaling an operator is another problem we don't need to worry about now. It's just something to keep in mind.

## Developer

Next I will:

- develop the algorithms for the user to choose from
- make a cute logo :)

### Algorithms and Actions needed...

Each reconcile will make a request to the queue and ask for updated information.
It will be on the endpoint (where flux is running) to store any state. Then the algorithn
selected by the user (run by the operator) must define conditions for:

- when to stop a MiniCluster (e.g., when is it done?)
- when to scale up
- when to scale down
- should there be an ability to ask for more jobs?
- Note that the _cluster_ autoscaler has a concept of [expanders](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler/expander) that can be tied to request nodes for specific pools. The more advanced setup of this operator will also have a cluster autoscaler.

Then test it out! We will want different kinds of scaling, both inside and outside. I think I know what I'm going to do and just need to implement it.
