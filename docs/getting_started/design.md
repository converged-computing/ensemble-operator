# Design

## Current Design

This current design moves the responsibility to monitor one or more ensembles from the operator to a service that is deployed alongside the members. It's a much better design to do that, and it also has a refactored [ensemble-python](https://github.com/converged-computing/ensemble-python) library that works external to the operator here (and Kubernetes in general) so you can run Ensembles alongside your workload manager a la carte!

![img/design.png](img/design.png)

More to come on this soon. Right now we have the ensemble running for Flux inside of Kubernetes, and I'll be adding grow capability soon.

## Design 1

This was the first design. I found it problematic to put the interaction between the operator and ensemble in a sidecar to the MiniCluster, because it meant I could not easily add another member. I also found the frequency of communication between the ensemble member and operator problematic - there was too much expectation or burden on the operator to receive and respond to requests. Finally, I didn't like that the Python library (that orchestrated the ensemble) was tied specifically to the operator. I wanted the library to work outside of it.

![img/design-1.png](img/design-1.png)

## 1. Create an Ensemble

We start with the creation of an Ensemble (1). If you look in [examples](https://github.com/converged-computing/ensemble-operator/tree/main/examples/algorithms/workload/demand) you will see one for lammps. An ensemble custom resource definition allows you to define the specification for an entire Flux MiniCluster, meaning that (when run in batch) it is going to kick off launching a bunch of jobs to the queue, and likely they will stack up (meaning the queue will be full). This is why we use the 🥞️ emoji freely - ensembles are like pancakes! Note that you can create more than one MiniCluster per ensemble, and this is intending to support being able to run ensembles on different resources. We also intend to support other types of operators (or Kubernetes abstractions) such as a JobSet, and each will likely have a different strategy for the scaling (for example, the Flux queue can be shared with a sidecar, and other abstractions don't even have queues, so likely we will just be twewaking the CRD definitions themselves). More on that later.

## 2. Create Ensemble Member

Each entry you define in the ensemble (right not, just MiniCluster) is called a Member. Logically, the first step of the operator
is to create the member, and wait until this resource is ready. This means that for the case of the Flux Operator, you need to have it installed in the cluster. The Flux Operator handles the creation, although because it's created _by_ the Ensemble operator, we technically own it. It's pretty neat how that works. :)

## 3. Flux Operator Logic

While explaining the Flux Operator is out of scope for here, I thought I would so it's clear why this works. The Flux Operator uses an init container (with a view of Flux) to dynamically add this entire Flux install to a shared empty directory volume. This means that any other container that connects to the volume can use Flux. It also means that if two containers are connected to the same empty directory, they both can interact with the _same_ Flux install and socket, meaning the same queue. This is immenesely important for our design! What the operator does is add a sidecar container to run alongside your application (e.g., LAMMPS) and
it has ready to go the Python server definition for the same GRPC that the operator is going to be a client for. This means that we can start the main application to launch jobs onto the queue, but then also start the sidecar GRPC service. We use the same exact Python install and Flux socket (in the shared view) and then expose the entire queue from the service. In layman's terms, the pod ip address provides an endpoint that will return a JSON dump of both queue and node metrics.  It's largely up to you, the user, how you want to submit jobs, or even if you want to have custom logic within your batch script to do that. There is a lot of cool stuff we can try.

## 4. GRPC Client to Ask for Updates

Once the MiniCluster is up and running (and the operator continues reconciling until it has an ip address to connect to associated with the pod, this works based on a selector for the job and the exact index that has the lead broker) we can create a GRPC client
from inside of the operator. It pings the sidecar container and asks for a status. The sidecar promptly delivers what it knows, the custom Python functions it has to show queue and node stats. The data looks like this:

```json
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

Note that this set of metadata provided can easily be expanded. These were the easy things to grab.

## 5. Algorithms

At this point, the user will have started the ensemble with some algorithm (an interface we have not implemented yet) and the operator will take the queue and node data, combine that with the user preference, and take an action. An algorithm should know under
what conditions to do the following:

- when to stop a MiniCluster (e.g., when is it done? Some other failure state condition?)
- when to scale up
- when to scale down
- when to ask for more jobs

For the last bullet, remember that we are connected to the running flux broker. We can easily define a set of commands in the specification for the algorithm, and then have some condition under which (in the response to the GRPC server running in the sidecar) we actually tell it to do something. For example, if we are running simulations and the queue is empty? We would send that information back to the operator, and the operator would see that it's algorithm instructs to submit more jobs when that happens, and it would send this signal back. The thing that is so cool about this is that there is really no limit to what we can do - we just need to decide. Likely the sidecar gRPC server can provide optional endpoints that provide functionality to interact with Flux in any way you can imagine (submit, save, start a new broker, something else?) and then the algorithm can decide which of those functions to use when it returns the response to a status request.

And yes, this does start to tip toe into state machine territory, we probably don't want to make it too complicated.

Finally, we can bring a cluster autoscaler into the picture. The _cluster_ autoscaler has a concept of [expanders](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler/expander) that can be tied to request nodes for specific pools. Since operators can easily serve scale endpoints, we likely can find a way to coordinate the MiniCluster scale request with the actual cluster scaling. If we have different node pools for different MiniCluster then it would be easy to assign based on expanders, but otherwise we will need to think. Likely there is a way and we just need to try it out. TLDR: The more advanced setup of this operator will also have a cluster autoscaler.

This is wicked! This is definitely my favorite under 24 hour operator I've produced. It's really cool 😎️
