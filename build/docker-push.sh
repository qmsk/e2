#!/bin/bash
echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin

docker tag qmsk/e2 qmsk/e2:${TRAVIS_TAG#v}

docker push qmsk/e2
docker push qmsk/e2:${TRAVIS_TAG#v}
