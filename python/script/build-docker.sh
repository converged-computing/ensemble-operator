#!/bin/bash

# Helper script to build docker images

docker build -t ghcr.io/converged-computing/ensemble-operator-api:rockylinux9-test -f docker/Dockerfile.rocky .
docker push ghcr.io/converged-computing/ensemble-operator-api:rockylinux9-test
