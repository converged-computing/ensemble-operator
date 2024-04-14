import argparse
import datetime
import json
import logging
import sys
from concurrent import futures

import grpc

import ensemble_operator.algorithm as algorithms
import ensemble_operator.defaults as defaults
import ensemble_operator.members as members
import ensemble_operator.metrics as m
from ensemble_operator.protos import ensemble_service_pb2
from ensemble_operator.protos import ensemble_service_pb2_grpc as api

# IMPORTANT: this only works with global variables if we have one ensemble gRPC per pod.
# We are anticipating storing some state here with the ensemble mamber since the
# operator should not be doing that.
cache = {}
metrics = None


def get_parser():
    parser = argparse.ArgumentParser(
        description="Ensemble Operator API Endpoint",
        formatter_class=argparse.RawTextHelpFormatter,
    )
    subparsers = parser.add_subparsers(
        help="actions",
        title="actions",
        description="actions",
        dest="command",
    )

    # Local shell with client loaded
    start = subparsers.add_parser(
        "start",
        description="Start the running server.",
        formatter_class=argparse.RawTextHelpFormatter,
    )
    start.add_argument(
        "--workers",
        help=f"Number of workers (defaults to {defaults.workers})",
        default=defaults.workers,
        type=int,
    )
    start.add_argument(
        "--port",
        help=f"Port to run application (defaults to {defaults.port})",
        default=defaults.port,
        type=int,
    )
    return parser


class EnsembleEndpoint(api.EnsembleOperatorServicer):
    """
    An EnsembleEndpoint runs inside the cluster.
    """

    def record_event(self, event):
        """
        A global log to keep track of state.
        """
        global cache
        if event not in cache:
            cache[event] = 0
        cache[event] += 1

    def get_event(self, event, default):
        """
        Get an event from the log
        """
        global cache
        return cache.get(event) or default

    def count_inactive_periods(self, increment, reset=False):
        """
        Keep a count of inactive jobs.

        Each time we cycle through, we want to check if the queue is active
        (or not). If not, we add one to the increment, because an algorithm can
        use this to determine a cluster termination status.
        """
        return self.increment_period("inactive_periods", increment, reset)

    def reset_counter(self, payload):
        """
        Reset a named counter
        """
        global cache
        payload = json.loads(payload)
        if isinstance(payload, str):
            payload = [payload]
        for key in payload:
            if key in cache:
                print(f"Resetting counter for {key}")
                cache[key] = 0

    def count_waiting_periods(self, current_waiting):
        """
        Count subsequent waiting periods that are == or greater than last count

        This is an indicator that the queue is not moving. We are interested to see
        if the number of waiting has changed. We would want to trigger scaling events,
        for example, if the number waiting does not change over some period.
        We return the number waiting, and waiting periods >= the last check.
        """
        global cache
        previous_waiting = cache.get("waiting")

        # If we don't have a previous value, it's akin to 0
        # And this is an increase in waiting
        if previous_waiting is None or (current_waiting >= previous_waiting):
            return self.increment_period("waiting_periods", 1, False)

        # If we get here, the current waiting is < previous waiting, so we reset
        return self.increment_period("waiting_periods", 0, True)

    def count_free_nodes_increasing_periods(self, nodes):
        """
        Given nodes, count number of free nodes, and increasing periods.
        """
        global cache
        current_free = nodes.get("node_free_count")
        previous_free = cache.get("free_nodes")

        # If we haven't set this before, or we have more free nodes
        # add a count of 1 to the period
        if previous_free is None or (current_free > previous_free):
            return self.increment_period("free_nodes", 1, False)

        # This means current free is less than the previous free
        # so the nodes are being used (and we reset)
        return self.increment_period("free_nodes", 0, True)

    def increment_period(self, key, increment, reset):
        """
        Given a counter in in the cache, increment it or reset
        """
        global cache
        if key not in cache:
            cache[key] = 0
        if increment:
            cache[key] += increment
        elif reset:
            cache[key] = 0
        return cache[key]

    def RequestStatus(self, request, context):
        """
        Request information about queues and jobs.
        """
        global cache
        global metrics

        print(context)
        print(f"Member type: {request.member}")

        # Record count of check to our cache
        self.record_event("status")

        # This will raise an error if the member type (e.g., minicluster) is not known
        member = members.get_member(request.member)

        # If the flux handle didn't work, this might error
        try:
            payload = member.status()
        except Exception as e:
            print(e)
            return ensemble_service_pb2.Response(
                status=ensemble_service_pb2.Response.ResultType.ERROR
            )

        # Prepare counts for the payload
        payload["counts"] = {}

        # Add the count of status checks to our payload
        payload["counts"]["status"] = self.get_event("status", 0)

        # Increment by 1 if we are still inactive, otherwise reset
        # note that we don't send over an actual inactive count, inactive here is the
        # period, largely because we don't need it. This isn't true for waiting
        increment, reset = member.count_inactive(payload["queue"])
        payload["counts"]["inactive"] = self.count_inactive_periods(increment, reset)

        # Increment by 1 if number waiting is the same or greater
        waiting_jobs = member.count_waiting(payload["queue"])
        payload["counts"]["waiting_periods"] = self.count_waiting_periods(payload["counts"])

        # This needs to be updated after so the cache has the previous waiting for the call above
        payload["counts"]["waiting"] = waiting_jobs

        # Finally, keep track of number of periods that we have free nodes increasing
        payload["counts"]["free_nodes"] = self.count_free_nodes_increasing_periods(payload["nodes"])

        # Always update the last timestamp when we do a status
        payload["metrics"] = metrics.to_dict()
        print(json.dumps(payload))

        return ensemble_service_pb2.Response(
            payload=json.dumps(payload),
            status=ensemble_service_pb2.Response.ResultType.SUCCESS,
        )

    def RequestAction(self, request, context):
        """
        Request an action is performed according to an algorithm.
        """
        print(f"Algorithm {request.algorithm}")
        print(f"Action {request.action}")
        print(f"Payload {request.payload}")

        # Assume first successful response
        status = ensemble_service_pb2.Response.ResultType.SUCCESS
        response = ensemble_service_pb2.Response()

        # The member primarily is directed to take the action
        member = members.get_member(request.member)
        if request.action == "submit":
            try:
                member.submit(request.payload)
            except Exception as e:
                print(e)
                response.status = ensemble_service_pb2.Response.ResultType.ERROR

        # Reset a counter, typically after an update event
        elif request.action == "resetCounter":
            try:
                self.reset_counter(request.payload)
            except Exception as e:
                print(e)
                response.status = ensemble_service_pb2.Response.ResultType.ERROR

        # This can give a final dump / view of job info
        elif request.action == "jobinfo":
            try:
                infos = member.job_info()
                if infos:
                    print(json.dumps(infos))
                    response.payload = infos
            except Exception as e:
                print(e)
                response.status = ensemble_service_pb2.Response.ResultType.ERROR

        return response


def serve(port, workers):
    """
    serve the ensemble endpoint for the MiniCluster
    """
    global metrics
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=workers))
    api.add_EnsembleOperatorServicer_to_server(EnsembleEndpoint(), server)
    server.add_insecure_port(f"[::]:{port}")
    print(f"🥞️ Starting ensemble endpoint at :{port}")

    # Kick off metrics collections
    metrics = m.Metrics()
    server.start()
    server.wait_for_termination()


def main():
    """
    Light wrapper main to provide a parser with port/workers
    """
    parser = get_parser()

    # If the user didn't provide any arguments, show the full help
    if len(sys.argv) == 1:
        help()

    # If an error occurs while parsing the arguments, the interpreter will exit with value 2
    args, _ = parser.parse_known_args()
    logging.basicConfig()
    serve(args.port, args.workers)


if __name__ == "__main__":
    main()
