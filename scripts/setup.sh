#!/bin/bash

set -x

IMAGE="python:3.9"
REGISTRY="localhost:5000"

# sudo soci index rm `sudo soci index ls -q`
sudo nerdctl rm -f `sudo nerdctl ps -aq`

sudo nerdctl run -d -p 5000:5000 --restart=always --name registry registry:2.7

sudo nerdctl pull registry.hub.docker.com/library/$IMAGE
sudo nerdctl pull ghcr.io/stargz-containers/$IMAGE-esgz

sudo nerdctl image tag registry.hub.docker.com/library/$IMAGE  $REGISTRY/$IMAGE
sudo nerdctl image tag ghcr.io/stargz-containers/$IMAGE-esgz  $REGISTRY/$IMAGE-esgz

sudo nerdctl image rm registry.hub.docker.com/library/$IMAGE
sudo nerdctl image rm ghcr.io/stargz-containers/$IMAGE-esgz

sudo nerdctl --insecure-registry push --snapshotter soci $REGISTRY/$IMAGE
sudo nerdctl push $REGISTRY/$IMAGE-esgz

sudo nerdctl image rm $REGISTRY/$IMAGE
sudo nerdctl image rm $REGISTRY/$IMAGE-esgz
