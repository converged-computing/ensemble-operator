# Grow and Shrink Example

The Flux MiniCluster offers an ability to grow and shrink, which we do with a trick to tell Flux it has more brokers than it does, and then we can dynamically change the size of the MiniCluster (an indexed job) by way of the scale endpoint. This is possible because of the Indexed Job, and because we've set the number of completions equal to the parallelism.

With this example, we first fire off a 10 second sleep job. When this finishes, we submit 5x an echo "hello world" job. When *each* of those jobs finish, we fire off 2x 100 second sleep job. There is a rule that says "if the queue wait time for the long sleep exceeds 5 seconds, grow." This is allowed to happen 5 times, with 2x backoff. In practice, this means we will scale up to allow 5 more pods, and with a delay of 2 checks in between, allowing the long sleeps to run. When everything is done, the ensemble is terminated. This does not lead to termination of the ensemble operator object, as it could be the case that other ensemble members are running. You should start by [installing the operator](https://converged-computing.org/ensemble-operator/getting_started/user-guide.html) and then applying the example:


```bash
kubectl apply -f ./ensemble.yaml
```

You'll start with one member;

```console
$ kubectl get pods
NAME                        READY   STATUS     RESTARTS   AGE
ensemble-0-0-wrjpg          0/1     Init:0/1   0          2s
ensemble-76b67f5f64-x2nbw   1/1     Running    0          2s
```

You can watch logs via the main MiniCluster container (the ensemble member):

```bash
kubectl logs ensemble-0-0-xxxx -f
```
```console
...
Found active groups {'sleep-long'}
💗 HEARTBEAT
Found active groups {'sleep-long'}
💗 HEARTBEAT
Found active groups {'sleep-long'}
💗 HEARTBEAT
Found active groups {'sleep-long'}
💗 HEARTBEAT
Found active groups {'sleep-long'}
💗 HEARTBEAT
Found active groups {'sleep-long'}
 => trigger count.sleep-long.success
   terminate ensemble session
broker.info[0]: rc2.0: ensemble run --kubernetes --executor minicluster --host 10.96.174.125 --port 50051 --name ensemble-0 /ensemble-entrypoint/ensemble.yaml Exited (rc=0) 210.8s
broker.info[0]: rc2-success: run->cleanup 3.51414m
broker.info[0]: cleanup.0: flux jobtap remove perilog 2>/dev/null || true Exited (rc=0) 0.1s
broker.info[0]: cleanup.1: flux queue stop --quiet --all --nocheckpoint Exited (rc=0) 0.1s
broker.info[0]: cleanup.2: flux cancel --user=all --quiet --states RUN Exited (rc=0) 0.1s
broker.info[0]: cleanup.3: flux queue idle --quiet Exited (rc=0) 0.1s
broker.info[0]: cleanup-success: cleanup->shutdown 0.377583s
broker.info[0]: children-complete: shutdown->finalize 82.6725ms
broker.info[0]: rc3.0: /opt/software/linux-rocky9-x86_64/gcc-11.4.1/flux-core-0.61.2-va52p5ph7ylanopgk7ranjgtjgc2wb5y/etc/flux/rc3 Exited (rc=0) 0.2s
broker.info[0]: rc3-success: finalize->goodbye 0.195435s
broker.info[0]: goodbye: goodbye->exit 0.050736ms
```

And eventually you will end up with pods 0 through 5 completed. Note that the ensemble service is not deleted, as it could be the case you have other members.

```console
$ kubectl get pods
NAME                       READY   STATUS      RESTARTS   AGE
ensemble-0-0-7mprb         0/1     Completed   0          6m44s
ensemble-0-1-jwgxj         0/1     Completed   0          5m35s
ensemble-0-2-8vt27         0/1     Completed   0          5m26s
ensemble-0-3-d8jr4         0/1     Completed   0          5m17s
ensemble-0-4-5v5pt         0/1     Completed   0          5m8s
ensemble-0-5-9djmh         0/1     Completed   0          4m59s
ensemble-f9745774f-k97fn   1/1     Running     0          6m44s
```

And you can look at the flux queue (while it's running, if you shell in) to see the jobs run:

```bash
 flux jobs -a 
       JOBID USER     NAME       ST NTASKS NNODES     TIME INFO
    ƒ629Bxnb root     sleep       R      1      1   59.98s ensemble-0-1
    ƒ61Z6FJf root     sleep       R      1      1   1.541m ensemble-0-1
    ƒ5zwWYYP root     sleep      CD      1      1   1.667m ensemble-0-0
    ƒ5zJSrVm root     sleep      CD      1      1   1.667m ensemble-0-0
    ƒ5ygs9jV root     sleep      CD      1      1   1.667m ensemble-0-0
    ƒ5y8FRXu root     sleep      CD      1      1   1.667m ensemble-0-0
    ƒ5xWfimd root     sleep      CD      1      1   1.667m ensemble-0-0
    ƒ5wx3za3 root     sleep      CD      1      1   1.667m ensemble-0-0
    ƒ5wMxH67 root     sleep      CD      1      1   1.667m ensemble-0-0
    ƒ5vkNaKq root     sleep      CD      1      1   1.667m ensemble-0-0
    ƒ5vBkr8F root     echo       CD      1      1   0.011s ensemble-0-0
    ƒ5ubf8eK root     echo       CD      1      1   0.010s ensemble-0-0
    ƒ5u33QSj root     echo       CD      1      1   0.010s ensemble-0-0
    ƒ5tSwgxo root     echo       CD      1      1   0.010s ensemble-0-0
    ƒ5srqyUs root     echo       CD      1      1   0.011s ensemble-0-0
     ƒQVTcyu root     sleep      CD      1      1   10.01s ensemble-0-0
```

Note that I captured the above while the last two jobs were finishing. You can delete the ensemble when you are done:

```bash
kubectl delete -f ./ensemble.yaml
```

When there are additional members the grpc service (the last running container that provides the endpoint to communicate with) can handle things like fair share, etc.

