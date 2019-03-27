#!/bin/bash

set -euo pipefail

NAME="${NAME:-contained.af}"
PKG="${PKG:-github.com/genuinetools/${NAME}}"
REGISTRY="${REGISTRY:-quay.io/dongsupark}"

DOCKER_FLAGS="--rm -i --disable-content-trust=true -v ${PWD}:/go/src/${PKG} --workdir /go/src/${PKG}"

docker run ${DOCKER_FLAGS} \
	"${REGISTRY}/${NAME}:dev" \
	uglifyjs --output frontend/js/contained.min.js --compress --mangle -- \
		frontend/js/xterm.js \
		frontend/js/fit.js \
		frontend/js/jquery-2.2.4.min.js \
		frontend/js/questions.js \
		frontend/js/main.js

docker run ${DOCKER_FLAGS} \
	"${REGISTRY}/${NAME}:dev" \
	sh -c 'cat frontend/css/normalize.css frontend/css/bootstrap.min.css frontend/css/xterm.css frontend/css/custom.css | cleancss -o frontend/css/contained.min.css'
