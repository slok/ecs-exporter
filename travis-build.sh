#!/bin/bash -e

AWS_REGION="us-east-1"
# Run vet
make vet

# Run tests
make test

# Push latest build
make push
