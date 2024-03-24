import collections
import importlib.util
import inspect
import os
import shutil
import sys

import ensemble_operator.utils as utils

try:
    import flux
    import flux.job
    import flux.resource
except ImportError:
    sys.exit("Cannot import flux. Please ensure that flux Python bindings are on the PYTHONPATH.")


# Keep a global handle so we make it just once
handle = flux.Flux()


def get_node_metrics():
    """
    Single function to get node metrics:

    core free count
    core up count
    node up count
    node free count
    """
    rpc = flux.resource.list.resource_list(handle)
    listing = rpc.get()
    return {
        "node_cores_free": listing.free.ncores,
        "node_cores_up": listing.up.ncores,
        "node_up_count": len(listing.up.nodelist),
        "node_free_count": len(listing.free.nodelist),
    }


def get_queue_metrics():
    """
    Update metrics for counts of jobs in the queue

    See https://github.com/flux-framework/flux-core/blob/master/src/common/libjob/job.h#L45-L53
    for identifiers.
    """
    jobs = flux.job.job_list(handle)
    listing = jobs.get()

    # Organize based on states
    states = [x["state"] for x in listing["jobs"]]
    counter = collections.Counter(states)

    # Lookup of state name to integer
    lookup = {
        "new": 1,
        "depend": 2,
        "priority": 4,
        "sched": 8,
        "run": 16,
        "cleanup": 32,
        "inactive": 64,
    }

    # This is how to get states
    counts = {}
    for stateint, count in counter.items():
        state = flux.job.info.statetostr(stateint)
        counts[state] = count
    for state in lookup:
        if state not in counts:
            counts[state] = 0
    return counts


# Organize metric functions by name
metrics = {
    "nodes": get_node_metrics,
    "queue": get_queue_metrics,
}
