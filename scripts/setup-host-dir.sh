#!/bin/sh

set -euo pipefail

# This script is used in `make run` for testing the project
# locally. As we run the application in a docker container itself,
# which acts as the host machine, we create a directory in the docker
# container that will be mounted into the nested container in the weak
# docker profile.

DOCKER_HOST="$(docker ps | grep ${REGISTRY}/docker:userns | cut -d ' ' -f1)"

docker exec -it ${DOCKER_HOST} /bin/mkdir -p /var/tmp/shared
