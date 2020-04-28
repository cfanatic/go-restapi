#!/usr/bin/env bash
#
# This script builds the application from source using Docker

# Clean setup
docker stop netchat
docker rm netchat
docker rmi netchat

# Build image
docker build -t netchat:latest -f ../Dockerfile ../

# Run container in combination with a MySQL container
# docker run --name netchat -d --network host --link mysql:db netchat:latest

# Run container in combination with a local instance of MySQL
docker run --name netchat -d --network host netchat:latest
