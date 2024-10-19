# Hello World Example

This is a simple hello world example. It will run a series of sleep jobs that trigger two groups of echos, and
then the termination action is triggered when the second group of echo jobs has a certain number of successes (2).
Debug mode is off, and you can set logging->debug to true to see more output from the queue (metrics, events, etc.)
You should start by [installing the operator](https://converged-computing.org/ensemble-operator/getting_started/user-guide.html) and then applying the example:


```bash
kubectl apply -f ./ensemble.yaml
```

The main objects created are an ensemble service deployment (not used yet here) and the ensemble member, a Flux MiniCluster:

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
broker.info[0]: rc1.0: running /opt/software/linux-rocky9-x86_64/gcc-11.4.1/flux-core-0.61.2-va52p5ph7ylanopgk7ranjgtjgc2wb5y/etc/flux/rc1.d/02-cron
broker.info[0]: rc1.0: /opt/software/linux-rocky9-x86_64/gcc-11.4.1/flux-core-0.61.2-va52p5ph7ylanopgk7ranjgtjgc2wb5y/etc/flux/rc1 Exited (rc=0) 0.8s
broker.info[0]: rc1-success: init->quorum 0.753456s
broker.info[0]: online: ensemble-0-0 (ranks 0)
broker.info[0]: quorum-full: quorum->run 0.100634s
 => trigger start
   submit sleep (name:sleep),(command:sleep 10),(count:1),(nodes:1)
 => trigger count.sleep.success
   submit echo (name:echo),(command:echo hello world),(count:5),(nodes:1)
 => trigger echo
   submit echo-again (name:echo-again),(command:echo hello world again),(count:2),(nodes:1)
 => trigger count.echo-again.success
   terminate ensemble session
broker.info[0]: rc2.0: flux submit -N 1 -n1 --quiet --watch ensemble run /ensemble-entrypoint/ensemble.yaml Exited (rc=0) 10.8s
broker.info[0]: rc2-success: run->cleanup 10.8239s
broker.info[0]: cleanup.0: flux jobtap remove perilog 2>/dev/null || true Exited (rc=0) 0.1s
broker.info[0]: cleanup.1: flux queue stop --quiet --all --nocheckpoint Exited (rc=0) 0.1s
broker.info[0]: cleanup.2: flux cancel --user=all --quiet --states RUN Exited (rc=0) 0.1s
broker.info[0]: cleanup.3: flux queue idle --quiet Exited (rc=0) 0.1s
broker.info[0]: cleanup-success: cleanup->shutdown 0.338158s
broker.info[0]: children-none: shutdown->finalize 0.037922ms
broker.info[0]: rc3.0: /opt/software/linux-rocky9-x86_64/gcc-11.4.1/flux-core-0.61.2-va52p5ph7ylanopgk7ranjgtjgc2wb5y/etc/flux/rc3 Exited (rc=0) 0.2s
broker.info[0]: rc3-success: finalize->goodbye 0.192969s
broker.info[0]: goodbye: goodbye->exit 0.049975ms
```

If we looked at the last state of the metrics (with debug on) we would see that there are 10 "echo-again" jobs run, and this is because we
trigger that group to run each time `job-finish` is triggered for echo (5 times) and echo-again is run 2x each group.

```console
{'variance': {'sleep-pending': Var: 0., 'sleep-duration': Var: 0., 'echo-pending': Var: 0., 'echo-duration': Var: 0., 'echo-again-pending': Var: 0., 'echo-again-duration': Var: 0.}, 'mean': {'sleep-pending': Mean: 0.023845, 'sleep-duration': Mean: 10.011414, 'echo-pending': Mean: 0.024413, 'echo-duration': Mean: 0.011502, 'echo-again-pending': Mean: 0.024439, 'echo-again-duration': Mean: 0.011559}, 'iqr': {'sleep-pending': IQR: 0., 'sleep-duration': IQR: 0., 'echo-pending': IQR: 0.000299, 'echo-duration': IQR: 0.000427, 'echo-again-pending': IQR: 0.000234, 'echo-again-duration': IQR: 0.000852}, 'max': {'sleep-pending': Max: 0.023845, 'sleep-duration': Max: 10.011414, 'echo-pending': Max: 0.025044, 'echo-duration': Max: 0.012155, 'echo-again-pending': Max: 0.024673, 'echo-again-duration': Max: 0.013035}, 'min': {'sleep-pending': Min: 0.023845, 'sleep-duration': Min: 10.011414, 'echo-pending': Min: 0.024102, 'echo-duration': Min: 0.011156, 'echo-again-pending': Min: 0.024139, 'echo-again-duration': Min: 0.010944}, 'mad': {'sleep-pending': MAD: 0., 'sleep-duration': MAD: 0., 'echo-pending': MAD: 0., 'echo-duration': MAD: 0.000427, 'echo-again-pending': MAD: 0.000103, 'echo-again-duration': MAD: 0.000206}, 'count': {'sleep': {'finished': Count: 1., 'success': Count: 1.}, 'echo': {'finished': Count: 5., 'success': Count: 5.}, 'echo-again': {'finished': Count: 10., 'success': Count: 10.}}}
```

This requires two important things in the ensemble.yaml:

- repetitions: is set to 5 for the job-finish event, otherwise it defaults to run (and you'd only get 2 echo again jobs, and it would not terminate)
- terminate: is set as an action on this metric!

Note that you can suppress the additional output of the MiniCluster by setting logging->quiet to true in the MiniCluster
member section. Also note that (currently with submit and quiet) it looks like we don't see output until the end. I am going
to work on this so it better streams (I thought watch would accomplish this). Also note that this example is not explicitly
for the ensemble operator - it will work outside of it just with flux! Super cool.