# Ensemble

> The CRD "Custom Resource Definition" defines an Ensemble

A CRD is a yaml file that you can apply to your cluster (with the Ensemble Operator
installed) to ask for a Ensemble. Kubernetes has these [custom resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) to make it easy to automate tasks, and in fact this is the goal of an operator!
A Kubernetes operator is conceptually like a human operator that takes your CRD,
looks at the cluster state, and does whatever is necessary to get your cluster state
to match your request. In the case of the Ensemble Operator, this means creating the resources
for an Ensemble. This document describes the spec of our custom resource definition.

## Custom Resource Definition

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

### EnsembleSpec

This is the top level of your CRD, the "spec" section below the header:

```yaml
apiVersion: ensemble.flux-framework.org/v1alpha1
kind: Ensemble
metadata:
  name: workload-demand
spec:
```

#### CheckSeconds

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

#### Algorithm

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

#### Members

Members is a list of members to add to your ensemble. In the future this could span different kinds of operators,
but for now we are focusing on Flux Operator MiniCluster, which has a nice setup to allow for a sidecar container
to monitor the Flux queue, doing everything from submitting jobs to reporting status. This is a list, so you
could have two MiniCluster types, for example, that have different resources. At some point we might want to
think about some kind of shared jobs matrix that can be defined at the top level of the ensemble, and have work
assigned to different ensemble members dynamically. For each member, you can define the following:

##### MiniCluster

Defining a Member.Minicluster spec assets that the member type is a MiniCluster (each member can only have one type).
This is the example we showed above. See [the MiniCluster definition](https://flux-framework.org/flux-operator/getting_started/custom-resource-definition.html)
for the various attributes. Note that the ensemble operator is not going to use any directive for a command, because it will always
start the MiniCluster in interactive mode.

##### Sidecar

The sidecar is where the gRPC service runs alongside the MiniCluster (or other member). You can customize options related
to this deployment, although you likely don't need to. I find this useful for development (e.g., using a development container
and asking to pull always). These are the options available:


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

Note that for sidecar images, we provide automated builds for two versions of each of rocky and ubuntu.
You can find them [here](https://github.com/converged-computing/ensemble-operator/pkgs/container/ensemble-operator-api).

##### Jobs

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


##### Member Algorithm

This directive is the same as the higher level [Algorithm](#algorithm), but can be set on the level of the member to override some global default.
