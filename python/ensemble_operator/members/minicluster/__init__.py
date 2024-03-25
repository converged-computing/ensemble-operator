import json
import multiprocessing
import random
import shlex

from ensemble_operator.members.base import MemberBase


class MiniClusterMember(MemberBase):
    """
    The MiniCluster member type

    Asking a MiniCluster member for a status means asking the flux queue.
    """

    def __init__(self):
        """
        Create a new minicluster handle
        """
        import flux

        self.handle = flux.Flux()

    def status(self):
        """
        Ask the flux queue (metrics) for a status
        """
        import ensemble_operator.members.minicluster.metrics as metrics

        # Prepare a payload to send back
        payload = {}

        # The payload is the metrics listing
        for name, func in metrics.metrics.items():
            payload[name] = func()

        return payload

    def count_inactive(self, queue):
        """
        Keep a count of inactive jobs.

        Each time we cycle through, we want to check if the queue is active
        (or not). If not, we add one to the increment, and this represents the number
        of subsequent inactive queue states we have seen. If we see the queue is active
        we reset the counter at 0. An algorithm can use this to determine a cluster
        termination status e.g., "terminate after inactive for N checks."
        Return the increment to the counter, plus a boolean to say "reset" (or not)
        """
        active_jobs = (
            queue["new"]
            + queue["run"]
            + queue["depend"]
            + queue["sched"]
            + queue["priority"]
            + queue["cleanup"]
        )
        if active_jobs == 0:
            return 1, False
        return 0, True

    def count_waiting(self, queue):
        """
        Keep a count of waiting jobs
        """
        return queue["new"] + queue["depend"] + queue["sched"] + queue["priority"]

    def submit(self, payload):
        """
        Receive the flux handle and StatusRequest payload to act on.
        """
        print(payload)
        payload = json.loads(payload)

        # Allow this to fail - it will raise a value error that will propogate back to the operator
        import flux.job

        # do we want to randomize?
        randomize = payload.get("randomize", True)

        # First prepare the entire set
        jobs = []
        for job in payload.get("jobs") or []:
            tasks = job.get("tasks") or 1

            # Node count cannot be greater than task count
            # Assume the sidecar is on the same node
            if tasks < job["nodes"]:
                tasks = multiprocessing.cpu_count()

            for _ in range(job.get("count", 0)):
                jobs.append(
                    {
                        "command": shlex.split(job["command"]),
                        "workdir": job.get("workdir"),
                        "nodes": job["nodes"],
                        "duration": job.get("duration") or 0,
                        "tasks": tasks,
                    }
                )

        # Do we want to randomize?
        if randomize:
            random.shuffle(jobs)

        # Now submit, likely randomized
        for job in jobs:
            print(job)
            jobspec = flux.job.JobspecV1.from_command(
                command=job["command"], num_nodes=job["nodes"], num_tasks=job["tasks"]
            )
            workdir = job["workdir"]

            # Do we have a working directory?
            if workdir:
                jobspec.cwd = workdir

            # Use direction or default to 0, unlimited
            jobspec.duration = job["duration"]

            # TODO do we want to customize environment somehow?
            jobid = flux.job.submit(self.handle, jobspec)
            print(f"  ⭐️ Submit job {job['command']}: {jobid}")
