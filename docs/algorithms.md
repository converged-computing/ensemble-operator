# Algorithm Design

The Ensemble Operator relies on a user-selected algorithm to decide on how to scale the queue, terminate a member, or even issue actions to a member (more advanced use cases). We will implement this design in three phases:

1. Reactive: the algorithm requests state, possibly updates that state to send back, and takes an action.
2. Controller: the algorithm requests state, updates state to send back, but also might provide an action for the queue to take.
3. Advanced: there is a machine learning model inside of the server that is trying to optimize some condition, and when a request for state is done, it calculates the desired action to send to the operator. The operator then assesses the current state it sees and decides to take the action (or not) sending this decision back to the sidecar.

We are going to start with the simplest design, and the details are discussed here. Status table:

| Algorithm | Status |
|-----------|------- |
| workload-demand | ✅️ implemented |
| select-fastest-runtime | 🟠️ not yet |
| select-longest-runtime | 🟠️ not yet |
| random-selection | 🟠️ not yet |
| workload-success | 🟠️ not yet |

## State

While operators are not intended to hold state (it is supposed to be represented in the cluster) I realized that there isn't any reason we cannot store some cache of state in the server running in the member sidecar directly. For example:

- An algorithm might have a rule that if the maximum jobs are running and the size of the queue has not dropped after N checks, that is an indication to scale the cluster by the size of the next smallest job. We might not have (directly in the operator) this state, but it can be sent over (on demand) for a check we are on, and then given a scaling decision, a response is sent back to indicate any action that the queue (and flux) should take.
- If we have a streaming ML server running in the sidecar, it might want to store a last state of a model to assess some change.
- I often refer to starting up a queue of jobs that aren't running (set priority to 0) and I think we should call this a queue in a frozen state. 🥶️

High level, we introduce the idea of state as a communication mechanism between the member (that holds state) and the operator (which does not) that must still make decisions that often require state. This means we need to define the valid actions for each algorithm, and ensure the logic is right so that the ensemble operator can make decisions relying on the member to hold its state.

## Objectives

I think we can simply say that any algorithm should be designed to meet some objective. For example:

- In many cases achieving the shortest time is likely what we want.
- We might also want to minimize cost
- A metric of goodness or stat derived from a model running in the sidecar gRPC service (e.g., stop when we have this accuracy, or this many successful simulations)

We have immense freedom in terms of what we can optimize, which is why designing these is pretty fun.

## Algorithms

Note that these are off the top of my head, and can be extended (or shrunk). Each can (and likely will be) extended with options, but I need to implement the core structure first.
Note that each algorithm will also define the member types (e.g., minicluster) that are supported.

### Structure

An algorithm choice is the selection (by the user) of a set of actions or rules for each of the following:

- scale up/down
- terminate
- change check frequency
- request an action by the cluster (advanced)

In the operator code, these actions (or rules) will map to different interfaces that can be selected programmatically (meaning in the CRD as a set of rules for a member). We will likely choose a reasonable default that looks something like "scale and terminate according to the workload needs." For the terminate rule, note that we will have a setting for an ensemble that says "Do not terminate, but exist at the smallest size so I can request more work from you quickly." This will actually be something we need to test - the con is that it is still taking up resources. The pro is that the cluster might have some state worth saving. For simplicity, we will start with the first two cases, and design simple algorithms that decide when to scale up and down.

### Algorithms

Each algorithm can define rules for scaling up, down, and termination, and (most) have a paired coordinated function on the server (e.g., to maintain a model or return state). Note that some scheduling algorithms are embedded, for example:

- If you want priority scheduling, Flux [should support](https://flux-framework.readthedocs.io/projects/flux-rfc/en/latest/spec_30.html#implementation) this. You can assume if the queue is first in first out, the jobs you put in first have highest priority.
- I don't see cases where we'd want to cut or (pre-empt?) jobs. If we use cloud resources to run something, it should be the case we see jobs through completion. Thus, round robin scheduling and shortest time reminaing don't make sense to me in this context.
- Advanced use cases might include multiple queues running within the operator, but I'm not ready to think about that yet.
- High level observation - many of these algorithms require fine grained control in how the jobs are submit (or not, just queued) so that needs to be easy, or minimally, examples provided for each.

Algorithms are organized by the following:

 - 🟦️ **basic** is typically a more static approach that responds retroactively to a queue
 - 🚗️ **model** is a more advanced algorithm that still responds to the queue, but relies on advanced logic running inside the ensemble member
 - 🕹️ **control** allows the operator to give higher level cluster feedback to the model or queue to better inform a choice, or simply takes this state into account (e.g., think fair-share across ensemble members or a cluster)


#### Workoad Demand (of consistent sizes)

> 🟦️ This algorithm assumes a first come, first serve submission (the queue is populated by the batch job) and the cluster resources are adapted to support the needs of the queue (not implemented yet).

This rule should be the default because it's kind of what you'd want for an autoscaling approach - you want the cluster resources to match the needs of the queue, where the needs of the queue are reflecting in the jobs in it, and can expand up to some maximum size and reduce down to some minimum size (1). This is first come, first serve approach, meaning that we assume the user has submit a ton of jobs to the queue in batch, and whichever are submit first are going to be run first.

- **scale up rule**: check the number of jobs waiting in the queue vs. the max size in the cluster. If the number waiting exceeds some threshold over N checks, increase the cluster size to allow for one more job, up to the max size.
- **scale down rule**: check the number of jobs running, waiting in the queue, and max size in the cluster. If the number of jobs waiting hits zero and remains at zero over N checks, decrease the size of the cluster down to the exact number needed that are running.
- **terminate rule**: check the number of jobs running and waiting. If this value remains 0 over N checks, the work is done and clean up. If there is a parameter set to keep the minicluster running at minimum operation, scale down to 1 node.

##### Options

Note that all options must be string or integer (no boolean). So for a boolean put "yes" or "no" instead.

| Name | Description | Default | Choices or Type |
| -----|---------|-----------|---------|
| randomize |   randomize based on _group_ of job |"yes" | "yes" "no"  (boolean) |
| terminateChecks | number of subsequent inactive checks to receive to determine termination status | 10 | integer |

Note that we likely want to randomize across ALL jobs but this isn't supported yet (but can be). I chose this strategy (sending as a group with a count)
because it's less data to send "over the wire" so to speak! This algorithm will use the identifier:

- workload-demand

#### Random selection

> 🟦️ This algorithm chooses jobs to run at random, and the queue retroactively responds.

This rule is intended to represent the outcome when jobs are selected to run at random, and the cluster adapts in size to meet that need.
This means that the user submits jobs that are not intended to run (they are set with a priority 0, I think that should work) and an algorithm running inside the ensemble member selects a random N to run each time around.

- **scale up rule**: check the number of jobs waiting in the queue vs. the max size in the cluster. If the number waiting exceeds some threshold over N checks, increase the cluster size to allow for one more job, up to the max size.
- **scale down rule**: check the number of jobs running, waiting in the queue, and max size in the cluster. If the number of jobs waiting hits zero and remains at zero over N checks, decrease the size of the cluster down to the exact number needed that are running.
- **terminate rule**: check the number of jobs running and waiting. If this value remains 0 over N checks, the work is done and clean up. If there is a parameter set to keep the minicluster running at minimum operation, scale down to 1 node.

This algorithm will use the identifier:

- random-selection


#### Workload Success

> 🕹️ Continue running and submitting jobs until a number of successful or valid is reached.

This algorithm assumes that we are running jobs (e.g., simulations) and many will fail, and we want to continue until we have reached a threshold of success.  This likely assumes jobs of the same type and size, but doesn't necessarily have to be. Since we don't know in advance how many jobs we need, we start the queue in a frozen state, and allow jobs to be selected as we go.

- **scale up rule**: check the number of jobs running, add resources (and request more submit) up to some maximum cluster size
- **scale down rule**: only scale down when we are at or equal to the number needed to be successful or valid. This is essentially allowing the running jobs to complete.
- **terminate rule**: terminate when we have the number of required or valid successful jobs and currently running jobs are complete.

This algorithm will use the identifier:

- workload-success


#### Select Fastest Runtime First

> 🚗️ A model based algorithm that selects work based on building a model of runtimes (not implemented yet)

This algorithm could be used for either workloads of different types (and different sizes) or a single workload that varies across sizes. The user submitting the job creates categories at the level that makes sense to build models for (workload type or size or maybe even both).
The Minicluster starts, but does not submit jobs. Instead, it adds them to the queue in some pending/ waiting state for a signal to start.

1. The jobs should be labeled in some manner, so it's clear they belong to a category (e.g., LAMMPS or AMG). A separate model will be built for each.
2. At initial start, there is no information about times, so one of each category is randomly selected.
3. The status keeps track of the number of completed for each type (e.g., LAMMPS 10, AMG 12)
4. At each update, the active queue is checked for which workloads are running.
  - The internal models are first updated. Jobs that are newly completed (and not known to the model) are used for training
  - If we have reached a minimum number of valid workflows to predict time for each, we use the model to select the next fastest to complete workload.
  - If we have not reached that value, we continue submitting "fairly" - at an equal distributed between each.

- **scale up rule**: when a selected workload (based on runtime) does not have enough resources, we scale up to the size that is needed to accommodate it.
- **scale down rule**: when a selected workload (based on runtime) has too many resources, we scale down.
- **terminate rule** when the queue is empty after N checks.

The above might be modified by adding another variable for a number of workloads allowed to run at once. E.g., we always select the max allowed to run - the current in the queue to select some new number of jobs,
and then choose the two with the fastest run time, and scale up or down accordingly.

This algorithm will use the identifier:

- select-fastest-runtime

This approach could arguably be improved by taking in a pre-existing model at the loading time, so we start all ready to go and don't need to wait to build the model. I am planning on using River for a streamling ML approach.


#### Select Longest Runtime First

> 🚗️ Model based algorithm to select work based

This algorithm is the same as the above, except we select for the longest runtimes first. In practice I'm not sure the use cases this might address, but the implication is that the longest runtimes are somehow more important.

- select-longest-runtime


#### Cost Based Selection

> 🕹️ Select an ensemble member to schedule to based on cost.

We know that folks aren't great at [connecting resource usage to costs](https://www.infoq.com/news/2024/03/cncf-finops-kubernetes-overspend/).
This setup would require multiple ensemble members, each associated with a node pool (that has a different instance type). For this kind of scaling, we have a goal to run a set of jobs across ensemble members, and the Ensemble Operator will want to trigger a job run (and scaling up or down) of resources based on cost estimates. For example, if a workload can run more quickly on a GPU resource (despite being more expensive) we would add a node to it, and likely scale down other ensemble members not being used. This approach is advanced and would require a model to be maintained across ensemble members. Likely we would have another CRD in the namespace for an `EnsembleModel` that keeps this metadata, and would be updated at each reconcile when a decision is warranted. This would also require starting up with a set of instance types to select from and costs - an added complexity that (while not impossible to provide) would make it harder to use the operator
overall.
