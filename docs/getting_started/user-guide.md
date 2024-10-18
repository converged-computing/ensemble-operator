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

And the ensemble operator:


```bash
kubectl apply -f  https://raw.githubusercontent.com/converged-computing/ensemble-operator/main/examples/dist/ensemble-operator.yaml
```

We also provide a build for ARM:

```bash
kubectl apply -f https://raw.githubusercontent.com/converged-computing/ensemble-operator/main/examples/dist/ensemble-operator-arm.yaml
```

For both of the above, we recommend pinning the image digest since development is actively happening. You can find the automated builds for both images [here](https://github.com/converged-computing/ensemble-operator/pkgs/container/ensemble-operator).

Or you can clone the repository and install locally.

```bash
git clone https://github.com/converged-computing/ensemble-operator
cd ensemble-operator
kubectl apply -f examples/dist/ensemble-operator.yaml
```

or to build and develop:

```bash
make test-deploy-recreate
```

Next let's talk about running the CRD, running LAMMPS!

### 2. Run Example

While we could hard code examples here, we recommend that you follow one of own included examples on GitHub.
Apologies this is a bit sparse at the moment - the project is new!

 - [workload-demand](https://github.com/converged-computing/ensemble-operator/tree/main/examples/algorithms/workload/demand) with LAMMPS.

Hopefully more coming soon! If you would like to request an example, please [let us know](https://github.com/converged-computing/ensemble-operator/issues).
