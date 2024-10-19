# The Ensemble Operator

Welcome to the Ensemble Operator Documentation!

The Ensemble Operator is a Kubernetes Cluster [Operator](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/) that you can install to your cluster to create and control ensembles of worklaods, and change them according to a user-specified ensemble, which is an algorithm dictated by a definition of work (jobs) and rules (triggers and actions). You can learn more via the links below.

 - [User Guide](getting_started/user-guide.md)
 - [Design](getting_started/design.md)
 - [Developer](developer/developer.md)

Since an entity in an ensemble is typically more complex than a container, while we currently just support a Flux Operator MiniCluster, we will eventually support a larger set of notable Kubernetes abstractions, including [Job](https://kubernetes.io/docs/concepts/workloads/controllers/job/), [JobSet](https://github.com/kubernetes-sigs/jobset), and [LeaderWorkersSet](https://github.com/kubernetes-sigs/lws).

**Important** This operator is in the midst of a refactor, and not production quality yet. Please come back soon.

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
