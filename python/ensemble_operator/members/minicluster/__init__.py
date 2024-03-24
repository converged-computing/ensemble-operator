import json
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

    def submit(self, payload):
        """
        Receive the flux handle and StatusRequest payload to act on.
        """
        print(payload)
        payload = json.loads(payload)

        # Allow this to fail - it will raise a value error that will propogate back to the operator
        import flux.job

        for job in payload.get("jobs") or []:
            print(job)
            command = shlex.split(job["command"])
            jobspec = flux.job.JobspecV1.from_command(command=command, num_nodes=job["nodes"])
            jobid = flux.job.submit(self.handle, jobspec)
            print(f"  ⭐️ Submit job {job['command']}: {jobid}")
