#!/bin/bash

cd ../ensemble
make kind-load
cd ../ensemble-operator
make test-deploy-recreate