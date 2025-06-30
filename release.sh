#!/bin/bash

echo "Build manifests and installer yaml files in local"
make generate manifests build bundle build-installer 

echo "Run lint to make sure it passed github action lint checks"
make lint

echo "Run unit test"
make test

# maybe check the env to run e2e before push


