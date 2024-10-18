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

  # The service deployment sidecar is where the ensemble sidecar gRPC service 
  # runs for all ensemble members in the space. It exists to remove the burden
  # off of the operator itself. 
  - sidecar:
      pullAlways: true
      image: ghcr.io/converged-computing/ensemble-python:latest

  # Each ensemble can hold one or more members.
  # The only valid member type is currently a Flux Operator MiniCluster
  members:

    # The algorithm is represented in the ensemble.yaml, which has a set of jobs
    # that are governed by rules (each with a trigger and action to take)
   - ensemble: |
      jobs:
        - name: lammps
          command: lmp -v x 2 -v y 2 -v z 2 -in in.reaxc.hns -nocite
          count: 10
          nodes: 1

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

#### Sidecar

The sidecar is where the gRPC service (deployment) runs alongside the members. You can customize options related
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

      # Change the container base (e.g., for arm)
      image: ghcr.io/converged-computing/ensemble-python:latest

      # The port for the gRPC service, I don't see why you'd need to change but maybe
      # This is the default.
      port: 50051

      # Number of workers to run for service (defaults to 10)
      workers: 10
```



#### Members

Members is a list of members to add to your ensemble. In the future this could span different kinds of operators,
but for now we are focusing on Flux Operator MiniCluster, which has a nice setup to allow for a sidecar container
to monitor the Flux queue, doing everything from submitting jobs to reporting status. This is a list, so you
could have two MiniCluster types, for example, that have different resources. For each member, you can define the following:

##### Ensemble

The ensemble section is a text chunk that should coincide with the ensemble.yaml that is described by ensemble-python. It will create a config map that is mapped as a volume to run the ensemble.


##### MiniCluster

Defining a Member.Minicluster spec assets that the member type is a MiniCluster (each member can only have one type).
This is the example we showed above. See [the MiniCluster definition](https://flux-framework.org/flux-operator/getting_started/custom-resource-definition.html)
for the various attributes. Note that the ensemble operator is not going to use any directive for a command, because it will always
start the MiniCluster in interactive mode.

Note that for sidecar images, we provide automated builds for two versions of each of rocky and ubuntu.
You can find them [here](https://github.com/converged-computing/ensemble-operator/pkgs/container/ensemble-operator-api).

