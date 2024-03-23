import argparse
import json
import logging
import sys
from concurrent import futures

import grpc

import ensemble_operator.defaults as defaults
from ensemble_operator.protos import ensemble_service_pb2
from ensemble_operator.protos import ensemble_service_pb2_grpc as api


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
        """Request information about queues and jobs."""
        print(request)
        print(context)

        # If the flux handle didn't work, this might error
        try:
            import ensemble_operator.metrics as metrics
        except:
            return ensemble_service_pb2.StatusResponse(
                status=ensemble_service_pb2.StatusResponse.ResultType.ERROR
            )

        # Prepare a payload to send back
        payload = {}

        # The payload is the metrics listing
        for name, func in metrics.metrics.items():
            payload[name] = func()

        print(json.dumps(payload, indent=4))
        # context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        # context.set_details('Method not implemented!')
        return ensemble_service_pb2.StatusResponse(
            payload=json.dumps(payload),
            status=ensemble_service_pb2.StatusResponse.ResultType.SUCCESS,
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
