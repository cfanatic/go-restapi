#!/usr/bin/env bash
#
# This script builds the application from source using Docker

docker build -t netchat:latest -f ../Dockerfile ../
docker run --name netchat -d -p 1025:1025 --link mysql:db netchat:latest
