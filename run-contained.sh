#!/bin/bash

set -euo pipefail

NAME="${NAME:-contained.af}"
PKG="${PKG:-github.com/genuinetools/${NAME}}"
REGISTRY="${REGISTRY:-quay.io/dongsupark}"

DOCKER_FLAGS="--rm -i -t --net=host --disable-content-trust=true --volume=/var/run/docker.sock:/var/run/docker.sock --volume=$PWD/.certs:/etc/docker/ssl:ro"

docker run ${DOCKER_FLAGS} \
	"${REGISTRY}/${NAME}:latest" -d \
	--dcacert=/etc/docker/ssl/cacert.pem \
	--dcert=/etc/docker/ssl/client.cert \
	--dkey=/etc/docker/ssl/client.key \
	--dhost=tcp://127.0.0.1:2375 \
	--port=10000

