#!/usr/bin/env bash

# export DOCKER_HOST=tcp://localhost:2376
export DOCKER_HOST=tcp://sysplex2.internal.logitbot.com:2376

REGISTRY="us.gcr.io/advisorconnect-1238/"
IMAGE="postgres"
VERSION="0.0.5"

BRANCH=`git branch | cut -c 3-`
COMMIT=`git log  | head -n1 | cut -c 8-16`

echo "Building ${REGISTRY}${IMAGE}:${VERSION}"
echo "BRANCH: $BRANCH"
echo "COMMIT: $COMMIT"

sed -l -e "s/BRANCH--/${BRANCH}/g" -e "s/COMMIT--/${COMMIT}/g" Dockerfile > Dockerfile.tmp

gcloud docker push -a
docker --tls build --pull --rm -t ${REGISTRY}${IMAGE}:${VERSION} -f Dockerfile.tmp .
docker --tls push ${REGISTRY}${IMAGE}:${VERSION}
docker --tls tag  ${REGISTRY}${IMAGE}:${VERSION} ${REGISTRY}${IMAGE}:latest
docker --tls push ${REGISTRY}${IMAGE}:${VERSION}

rm -f Dockerfile.tmp