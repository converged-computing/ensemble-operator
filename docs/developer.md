# Developer

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
