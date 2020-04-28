#!/usr/bin/env bash
#
# This script builds the application from source using Docker

# Build image
docker build -t netchat:latest -f ../Dockerfile ../

# Run container in combination with a MySQL container
# docker run --name netchat -d -p 1025:1025 --link mysql:db netchat:latest

# Run container in combination with a local instance of MySQL
docker run --name netchat -d -p 1025:1025 netchat:latest
