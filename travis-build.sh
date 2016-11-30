#!/bin/bash -e

# Run vet
make vet

# Run tests
make test

# Push latest build
make push
