# Workload Demand Experiment

You can read about the workload demand algorithm [here](https://github.com/converged-computing/ensemble-operator/blob/main/docs/algorithms.md#workoad-demand-of-consistent-sizes). Here we are doing a small experiment to test the following cases:

- independent base case without ensemble (launching separate Miniclusters for each job) at max size
- single minicluster with all jobs in queue at max size needed
- single minicluster with all jobs in queue at min size needed
- single minicluster with all jobs with autoscaling
- workload driven ensemble with autoscaler enabled and different submit approaches
  - random submit
  - ascending job size
  - descending job size

## 1. Create the Cluster

```bash
GOOGLE_PROJECT=myproject
gcloud container clusters create test-cluster \
    --threads-per-core=1 \
    --placement-type=COMPACT \
    --num-nodes=1 \
    --region=us-central1-a \
    --project=${GOOGLE_PROJECT} \
    --machine-type=c2d-standard-8 \
    --enable-autoscaling --total-min-nodes=1 --total-max-nodes=6 \
    --autoscaling-profile optimize-utilization
```

Note that there are two [autoscaling profiles](https://cloud.google.com/kubernetes-engine/docs/concepts/cluster-autoscaler#autoscaling_profiles) balanced (default) and optimize-utilization. I first tested balanced and found
that nodes hung around ~10 minutes after the queue was essentially empty, so I think the second one (that is noted to be more
aggressive) might be better.

We are going to use the Flux operator for ensemble members, and the ensemble operator for ensembles.
We will enable autoscaling.

```bash
kubectl apply -f https://raw.githubusercontent.com/flux-framework/flux-operator/main/examples/dist/flux-operator.yaml
kubectl apply -f examples/dist/ensemble-operator.yaml
```

## 2. Understand how it works

Here is what is going to happen:

1. We define our jobs matrix in the [ensemble.yaml](ensemble.yaml). The jobs matrix is consistent across algorithms.
2. For any algorithm, we first get a cluster status `StatusRequest`. We don't use it here aside from establishing communication. In the future if we return more metadata about the cluster it can be used to inform decision making.
3. We then detect that the job matrix has outstanding jobs, and make an `ActionRequest` to "submit" the jobs.
4. The jobs are submit on the MiniCluster, and the matrix is emptied.
5. We proceed to monitor, scaling when conditions are met, downsizing when jobs are finishing, and terminating after that.

## 3. Create and run the ensemble

Create the job. Note that this [ensemble.yaml](ensemble.yaml), for this algorithm type, requires you to submit jobs. It is retroactively going to adjust the queue based on what you submit. We also are going to rely on the algorithm to determine when to terminate the minicluster, so we add a sleep command to the end. Arguably this could be controlled (added) by the operator based on seeing the algorithm type, but this works for now.

```bash
kubectl apply -f ensemble.yaml
```

We can check both the gRPC sidecar and the operator to see if information is being received. Here is the
sidecar (after setup steps):

```bash
kubectl logs workload-demand-0-0-vfxxd -c api -f
```
```console
🥞️ Starting ensemble endpoint at :50051
<grpc._server._Context object at 0x7f41ab4161d0>
Member type: minicluster
{"nodes": {"node_cores_free": 10, "node_cores_up": 10, "node_up_count": 1, "node_free_count": 1}, "queue": {"new": 0, "depend": 0, "priority": 0, "sched": 0, "run": 0, "cleanup": 0, "inactive": 0}, "counts": {"status": 1, "inactive": 1}}
Algorithm workload-demand
Action submit
Payload {"jobs":[{"name":"lammps-2","command":"lmp -v x 2 -v y 2 -v z 2 -in in.reaxc.hns -nocite","count":10,"nodes":1},{"name":"lammps-4","command":"lmp -v x 4 -v y 4 -v z 4 -in in.reaxc.hns -nocite","count":5,"nodes":1}]}
{"jobs":[{"name":"lammps-2","command":"lmp -v x 2 -v y 2 -v z 2 -in in.reaxc.hns -nocite","count":10,"nodes":1},{"name":"lammps-4","command":"lmp -v x 4 -v y 4 -v z 4 -in in.reaxc.hns -nocite","count":5,"nodes":1}]}
{'name': 'lammps-2', 'command': 'lmp -v x 2 -v y 2 -v z 2 -in in.reaxc.hns -nocite', 'count': 10, 'nodes': 1}
  ⭐️ Submit job lmp -v x 2 -v y 2 -v z 2 -in in.reaxc.hns -nocite: ƒ6PTddpB
  ⭐️ Submit job lmp -v x 2 -v y 2 -v z 2 -in in.reaxc.hns -nocite: ƒ6Q8BK99
  ⭐️ Submit job lmp -v x 2 -v y 2 -v z 2 -in in.reaxc.hns -nocite: ƒ6QpCykT
  ⭐️ Submit job lmp -v x 2 -v y 2 -v z 2 -in in.reaxc.hns -nocite: ƒ6RUkf5R
  ⭐️ Submit job lmp -v x 2 -v y 2 -v z 2 -in in.reaxc.hns -nocite: ƒ6S7pM83
  ⭐️ Submit job lmp -v x 2 -v y 2 -v z 2 -in in.reaxc.hns -nocite: ƒ6Skt3Af
  ⭐️ Submit job lmp -v x 2 -v y 2 -v z 2 -in in.reaxc.hns -nocite: ƒ6TPwjDH
  ⭐️ Submit job lmp -v x 2 -v y 2 -v z 2 -in in.reaxc.hns -nocite: ƒ6U1XRyZ
  ⭐️ Submit job lmp -v x 2 -v y 2 -v z 2 -in in.reaxc.hns -nocite: ƒ6Ud78jq
  ⭐️ Submit job lmp -v x 2 -v y 2 -v z 2 -in in.reaxc.hns -nocite: ƒ6VEgqW7
{'name': 'lammps-4', 'command': 'lmp -v x 4 -v y 4 -v z 4 -in in.reaxc.hns -nocite', 'count': 5, 'nodes': 1}
  ⭐️ Submit job lmp -v x 4 -v y 4 -v z 4 -in in.reaxc.hns -nocite: ƒ6VrGYGP
  ⭐️ Submit job lmp -v x 4 -v y 4 -v z 4 -in in.reaxc.hns -nocite: ƒ6WTrF2f
<grpc._server._Context object at 0x7f41a255de90>
Member type: minicluster
{"nodes": {"node_cores_free": 0, "node_cores_up": 10, "node_up_count": 1, "node_free_count": 0}, "queue": {"sched": 5, "run": 10, "new": 0, "depend": 0, "priority": 0, "cleanup": 0, "inactive": 0}, "counts": {"status": 2, "inactive": 0}}
```

You'll see the jobs received and submit, and then after we won't call that again because the matrix is empty, but it will be followed by subsequent calls to see the queue (the last line).
Here is what the operator sees at the start. I am using non-traditional logging here because I really don't like the standard json formatting that everyone uses - it provides a ton of information that I don't need and I can't find what I'm looking for.
I think this is better with organized sections.

```bash
kubectl logs -n ensemble-operator-system ensemble-operator-controller-manager-5f874bb7d8-m68jb 
```
```console
🥞️ Ensemble! Like pancakes
   => Request: default/workload-demand

🤓 Ensemble.members 1
   => Ensemble.member 0
      Ensemble.member.Algorithm: workload-demand
      Ensemble.member Type: minicluster
      Ensemble.member.Sidecar.Image: ghcr.io/converged-computing/ensemble-operator-api:rockylinux9-test
      Ensemble.member.Sidecar.Port: 50051
      Ensemble.member.Sidecar.PullAlways: true
      Members 1
✨ Ensuring Ensemble MiniCluster
      Found existing Ensemble MiniCluster
      Checking member workload-demand-0
🦀 MiniCluster Ensemble Update
      Pod IP Address 10.244.1.59
      Host 10.244.1.59:50051

...

Member minicluster has active jobs or has not met threshold for for termination
```

What you'll see from the operator is we are doing scale up and down and termination checks based on the number of subsequent inactive statuses. We will want to see a threshold reached (a small one here, just 2) before the cluster
is terminated.  As long as something is in states:

- run
- new
- depend
- cleanup
- priority
- sched

It is considered active, and the inactive count will not increment. When we hit conditions for scaling, you'll see the MiniCluster get larger, and when jobs are done (and we determine inactive status) you'll see this happen in the operator:

```console
Member minicluster is marked for termination
SUCCESS
2024-03-24T16:36:15Z    INFO          Ensemble is Ready!        {"controller": "ensemble", "controllerGroup": "ensemble.flux-framework.org", "controllerKind": "Ensemble", "Ensemble": {"name":"workload-demand","namespace":"default"}, "namespace": "default", "name": "workload-demand", "reconcileID": "8fa68ccc-7ffe-453a-ace9-a0532d78d228"}
🥞️ Ensemble! Like pancakes
   => Request: default/workload-demand
      Ensemble not found. Ignoring since object must be deleted.
```

On the side of the worker (lead broker in the gRPC sidecar), you'll see the same - it report an inactive count greater than the threshold:

```console
{"nodes": {"node_cores_free": 20, "node_cores_up": 20, "node_up_count": 2, "node_free_count": 2}, "queue": {"inactive": 15, "new": 0, "depend": 0, "priority": 0, "sched": 0, "run": 0, "cleanup": 0}, "counts": {"status": 43, "inactive": 3}}
Algorithm workload-demand
Action terminate
Payload 
```

And will exit cleanly (pods will be terminated and go away). Note that by default, we randomize the submission of the original jobs. The [ensemble.yaml](ensemble.yaml) modifies that to 2 to make it faster. These
options are customizable with the algorithm. 

## 4. Run equivalent miniclusters

At this point we could launch the directory of [minicluster](minicluster) to do exactly that, and compare some attributes between the two cases. Would we compare total runtime? Costs? Is the assumption that the autoscaler is running? Need to chat about what experiments to do here - I think the operator is working.


## 5. Things to Talk about

- In the base case, we can run the workload autoscaling, and give jobs to be run in some order. The ensemble operator will see the needs of the queue, and then react to size the cluster to that. Thus, we can vary the ordering of that in some way that makes sense.
- We can also integrate priority here, and have all the jobs submit with a priority that won't run, and then have something run in the server (that works with the emsemble operator) to submit jobs based on its own internal logic.
- Either way, we likely want to compare to a base case of running all jobs on the smallest sized cluster, and the largest sized cluster, and accounting for time as a proxy for cost. We also probably 
