import argparse
import json
import logging
import sys
from concurrent import futures

import grpc

import ensemble_operator.algorithm as algorithms
import ensemble_operator.defaults as defaults
import ensemble_operator.members as members
from ensemble_operator.protos import ensemble_service_pb2
from ensemble_operator.protos import ensemble_service_pb2_grpc as api

# IMPORTANT: this only works with global variables if we have one ensemble gRPC per pod.
# We are anticipating storing some state here with the ensemble mamber since the
# operator should not be doing that.
cache = {}


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

    def count_inactive(self, increment, reset=False):
        """
        Keep a count of inactive jobs.

        Each time we cycle through, we want to check if the queue is active
        (or not). If not, we add one to the increment, because an algorithm can
        use this to determine a cluster termination status.
        """
        global cache
        if "count_inactive" not in cache:
            cache["count_inactive"] = 0
        if increment:
            cache["count_inactive"] += increment
        elif reset:
            cache["count_inactive"] = 0
        return cache["count_inactive"]

    def RequestStatus(self, request, context):
        """
        Request information about queues and jobs.
        """
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

        # Add the count of status checks to our payload
        status_count = self.get_event("status", 0)
        # Increment by 1 if we are still inactive, otherwise reset
        increment, reset = member.count_inactive(payload["queue"])
        inactive_count = self.count_inactive(increment, reset)
        payload["counts"] = {"status": status_count, "inactive": inactive_count}

        # Retrieve the algorithm to process the request.
        # TODO this can do additional logic / parsing, not being used yet
        # alg = algorithms.get_algorithm(request.algorithm)

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

        # The member primarily is directed to take the action
        member = members.get_member(request.member)
        if request.action == "submit":
            try:
                member.submit(request.payload)
            except Exception as e:
                print(e)
                status = ensemble_service_pb2.Response.ResultType.ERROR
        return ensemble_service_pb2.Response(status=status)


def serve(port, workers):
    """
    serve the ensemble endpoint for the MiniCluster
    """
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=workers))
    api.add_EnsembleOperatorServicer_to_server(EnsembleEndpoint(), server)
    server.add_insecure_port(f"[::]:{port}")
    print(f"🥞️ Starting ensemble endpoint at :{port}")
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
