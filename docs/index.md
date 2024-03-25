# The Ensemble Operator

Welcome to the Ensemble Operator Documentation!

The Ensemble Operator is a Kubernetes Cluster [Operator](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)
that you can install to your cluster to create and control ensembles of worklaods, and change them according to a user-specified algorithm. You can learn more via the links below.

 - [User Guide](getting_started/user-guide.md)
 - [Algorithms](getting_started/algorithms.md)
 - [Design](getting_started/design.md)
 - [Developer](developer/developer.md)

Since an entity in an ensemble is typically more complex than a container, while we currently just support a Flux Operator MiniCluster, we could eventually support a larger set of notable Kubernetes abstractions, including [Job](https://kubernetes.io/docs/concepts/workloads/controllers/job/), [JobSet](https://github.com/kubernetes-sigs/jobset), and [LeaderWorkersSet](https://github.com/kubernetes-sigs/lws).

These seem like a well-scoped set to start. For JobSet, LeaderWorkerSet, and MiniCluster, the corresponding operator for each is required to be installed to your cluster
to use them.

If you'd like to ask a question or contribute, please visit the repository [on GitHub](https://github.com/converged-computing/ensemble-operator).

```{toctree}
:caption: Getting Started
:maxdepth: 2
getting_started/index.md
developer/index.md
```

```{toctree}
:caption: About
:maxdepth: 2
about/index.md
```

<script>
// This is a small hack to populate empty sidebar with an image!
document.addEventListener('DOMContentLoaded', function () {
    var currentNode = document.querySelector('.md-sidebar__scrollwrap');
    currentNode.outerHTML =
	'<div class="md-sidebar__scrollwrap">' +
		'<img style="width:100%" src="_static/images/logo.png"/>' +

	'</div>';
}, false);

</script>
