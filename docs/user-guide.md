# User Guide

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

### 2. Custom Resource Definition

The custom resource definition is a YAML file that holds the options you can define. Let's look at an example for running LAMMPS
with the currently available algorithm "workload-demand" - the file is annotated to provide detail to each section, which is also provided below.

```yaml
apiVersion: ensemble.flux-framework.org/v1alpha1
kind: Ensemble
metadata:
  name: workload-demand
spec:

  # Each ensemble can hold one or more members.
  # The only valid member type is currently a Flux Operator MiniCluster
  members:

  # The member sidecar is where the ensemble sidecar gRPC service will be running
  # For the MiniCluster, this sidecar will have full access to the Flux queue
  - sidecar:
      pullAlways: true
      image: ghcr.io/converged-computing/ensemble-operator-api:rockylinux9-test

    # The algorithm you choose for a member determines behavior for scaling, termination,
    # and even submitting jobs. Each algorithm has its own set of options
    algorithm:
      name: workload-demand
      options:
        disableTermination: "yes"

    # This is your jobs matrix. These are submit by the ensemble operator via gRPC,
    # and the way that is done also depends on the algorithm
    jobs:
      - name: lammps-2
        command: lmp -v x 2 -v y 2 -v z 2 -in in.reaxc.hns -nocite
        count: 10
      - name: lammps-4
        command: lmp -v x 4 -v y 4 -v z 4 -in in.reaxc.hns -nocite
        count: 2

    # This is a full MiniCluster CRD, directly from the Flux Operator!
    # If you don't set a maxSize it defaults to size
    minicluster:
      spec:
        size: 2
        maxSize: 10
        minSize: 2

        # This is a list because a pod can support multiple containers
        containers:
        - image: ghcr.io/converged-computing/metric-lammps:latest

          # You can set the working directory if your container WORKDIR is not correct.
          workingDir: /opt/lammps/examples/reaxff/HNS
```

Now we will discuss each section in detail.

#### EnsembleSpec

This is the top level of your CRD, the "spec" section below the header:

```yaml
apiVersion: ensemble.flux-framework.org/v1alpha1
kind: Ensemble
metadata:
  name: workload-demand
spec:
```

##### CheckSeconds

After the creation of an ensemble member, we are going to do some kind of check every N seconds. While this could
be set on the level of the member, for the time being I am putting it at a more global level, meaning that every
member is checked at the same frequency. This could be changed.

```yaml
apiVersion: ensemble.flux-framework.org/v1alpha1
kind: Ensemble
metadata:
  name: workload-demand
spec:
  members:
    ...
  checkSeconds: 20
```

It could also be that we have (in an algorithm) an ability to change this check as we go, based on some conditions.

##### Algorithm

To select a global algorithm for all members, it can be defined here. Or you can define an algorithm for some members,
and then those that don't have one defined use the global one as a default. Each algorithm has its own custom options,
but here is an example:

```yaml
spec:
  algorithm:
    name: workload-demand
    options:
      scaleUpStrategy: "randomJob"
```

See [algorithms](algorithms.md) for more details about algorithms available and options for them.

##### Members

Members is a list of members to add to your ensemble. In the future this could span different kinds of operators,
but for now we are focusing on Flux Operator MiniCluster, which has a nice setup to allow for a sidecar container
to monitor the Flux queue, doing everything from submitting jobs to reporting status. This is a list, so you
could have two MiniCluster types, for example, that have different resources. At some point we might want to
think about some kind of shared jobs matrix that can be defined at the top level of the ensemble, and have work
assigned to different ensemble members dynamically. For each member, you can define the following:

###### MiniCluster

Defining a Member.Minicluster spec assets that the member type is a MiniCluster (each member can only have one type).
This is the example we showed above. See [the MiniCluster definition](https://flux-framework.org/flux-operator/getting_started/custom-resource-definition.html)
for the various attributes. Note that the ensemble operator is not going to use any directive for a command, because it will always
start the MiniCluster in interactive mode.

###### Sidecar

The sidecar is where the gRPC service runs alongside the MiniCluster (or other member). You can customize options related
to this deployment, although you likely don't need to. I find this useful for development (e.g., using a development container
and asking to pull always). These are the options available:


```yaml
```yaml
apiVersion: ensemble.flux-framework.org/v1alpha1
kind: Ensemble
metadata:
  name: workload-demand
spec:
  members:
  - sidecar:
      # Always pull the container
      pullAlways: true

      # Change the container base (e.g., for a different operator system)
      image: ghcr.io/converged-computing/ensemble-operator-api:rockylinux9-test

      # The port for the gRPC service, I don't see why you'd need to change but maybe
      # This is the default, and it is a string
      port: "50051"

      # Number of workers to run for service (defaults to 10)
      workers: 10
```

###### Jobs

Each member defines jobs to be submit in the jobs matrix. We do this because (early on) I [realized](https://github.com/converged-computing/ensemble-operator/issues/6)
requiring the user to write batch jobs with the correct setup to ask for waitable, and also to control sleeping of a worker or leader node, was really complex. A much
simpler and better controller strategy (and one that works well with algorithms that want to control submission) is just to control all of it. We start the MiniCluster
in interactive mode, and all requests to submit jobs come in directly from the ensemble operator. This is where the jobs matrix comes in - it is parsed and unwrapped into
your jobs to run.

```yaml
apiVersion: ensemble.flux-framework.org/v1alpha1
kind: Ensemble
metadata:
  name: workload-demand
spec:

  # Each ensemble can hold one or more members.
  # The only valid member type is currently a Flux Operator MiniCluster
  members:

    # This is your jobs matrix. These are submit by the ensemble operator via gRPC,
    # and the way that is done also depends on the algorithm
    jobs:
      - name: lammps-2
        command: lmp -v x 2 -v y 2 -v z 2 -in in.reaxc.hns -nocite
        count: 10
      - name: lammps-4
        command: lmp -v x 4 -v y 4 -v z 4 -in in.reaxc.hns -nocite
        count: 2
```

In the above, we see three named groups of jobs. You can decide how you want to group them - it could be based on an application,
an application and problem size, or something else. When we have algorithms that train ML models we will likely need groups that are categories
for training (e.g., predicting a runtime for a specific lammps problem size). For now, these groups are unwrapped and submit as needed,
depending on the algorithm. For example, the reactive "workload-demand" algorithm is going to expand the matrix, randomly shuffle the jobs,
and submit them all at the onset.  The variables should be intuitive - not shown are `workdir` that determines the working directory
for the work, and `duration` that can set a limit for something to run.


###### Member Algorithm

This directive is the same as the higher level [Algorithm](#algorithm), but can be set on the level of the member to override some global default.

Next let's talk about running the CRD, running LAMMPS!

### 3. Run LAMMPS

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

### 4. Check GRPC Service Endpoint

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
