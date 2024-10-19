# Ensemble

> The CRD "Custom Resource Definition" defines an Ensemble

A CRD is a yaml file that you can apply to your cluster (with the Ensemble Operator
installed) to ask for a Ensemble. Kubernetes has these [custom resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) to make it easy to automate tasks, and in fact this is the goal of an operator!
A Kubernetes operator is conceptually like a human operator that takes your CRD,
looks at the cluster state, and does whatever is necessary to get your cluster state
to match your request. In the case of the Ensemble Operator, this means creating the resources
for an Ensemble. This document describes the spec of our custom resource definition.

## Custom Resource Definition

The custom resource definition is a YAML file that holds the options you can define. Let's look at the hello-world example that will submit different groups of jobs. Specifically:

1. Submit a sleep job when the ensemble starts.
2. When the sleep job is successful, submit 5 echo jobs
3. Submit echo-again each time we have a single echo job finish (this means the rule will trigger 5 times)
4. When we have 10 echo-again finish (since the above is 5*2) terminate!

Note that if you don't have a terminate action, the ensemble will not terminate. If you don't specify repetitions for a job event (e.g., `job-finish`) it assumes 1.

```yaml
apiVersion: ensemble.flux-framework.org/v1alpha1
kind: Ensemble
metadata:
  name: ensemble
spec:  

  # This is how you change the sidcar image, if needed. This is the one
  # that I push and use for development. Pull always ensures we get latest
  sidecar:
    pullAlways: true
    image: ghcr.io/converged-computing/ensemble-python:latest

  members:

  # The MiniCluster that will run the ensemble
  - minicluster:
      spec:
        size: 1
        minSize: 1
        maxSize: 16
        
        # uncomment to make interactive and debug
        # interactive: true

        # The workers should not fail when they clean up
        flux:
          completeWorkers: true
        
        # This is a list because a pod can support multiple containers
        containers:
        - image: ghcr.io/converged-computing/metric-lammps:latest

          # You can set the working directory if your container WORKDIR is not correct.
          workingDir: /opt/lammps/examples/reaxff/HNS
          resources:
            limits:
              cpu: 3
            requests:
              cpu: 3

    # The ensemble configuration file (what is provided to ensemble python)
    ensemble: |
      logging:
        debug: true

      jobs:
        - name: echo
          command: echo hello world
          count: 5
          nodes: 1
        - name: echo-again
          command: echo hello world again
          count: 2
          nodes: 1
        - name: sleep
          command: sleep 10
          count: 1
          nodes: 1

      rules:
        # 1. This rule says to submit the sleep jobs when we start
        - trigger: start
          action:
            name: submit
            label: sleep

        - trigger: metric
          name: count.sleep.success
          when: 1
          action:
            name: submit
            label: echo

        # When we have a successful echo job, submit echo again
        # Do this 5 times (repetitions), otherwise it will just run once
        - trigger: job-finish
          name: echo
          action:
            name: submit
            label: echo-again
            repetitions: 5

        # Terminate the ensemble when we have 2 successful of echo-again
        # If you don't terminate the ensemble, the queue will keep waiting
        - trigger: metric
          name: count.echo-again.success
          when: 10
          action:
            name: terminate
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
