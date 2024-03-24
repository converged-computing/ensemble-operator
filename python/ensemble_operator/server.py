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

    def RequestStatus(self, request, context):
        """
        Request information about queues and jobs.
        """
        print(context)
        print(f"Member type: {request.member}")

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

        # TODO retrieve the algorithm to process the request
        # We aren't using this or doing that yet, we are just submitting jobs
        # alg = algorithms.get_algorithm(request.algorithm)
        member = members.get_member(request.member)
        if request.action == "submit":
            member.submit(request.payload)

        return ensemble_service_pb2.Response(
            status=ensemble_service_pb2.Response.ResultType.SUCCESS
        )


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
