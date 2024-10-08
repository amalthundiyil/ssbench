#!/bin/bash

set -xe

IMAGES=(
    "library/python:3.9" 
    "library/gcc:11.2.0"
    "clelange/cms-higgs-4l-full:latest"
    "rootproject/root:6.32.02-ubuntu24.04"
    "library/centos:7"
)

REGISTRY="localhost:5000"

if [[ $(sudo nerdctl ps -aq) ]]; then
    sudo nerdctl rm -f $(sudo nerdctl ps -aq)
fi

sudo nerdctl run -d -p 5000:5000 --restart=always --name registry registry:2.7 || true

for IMAGE in "${IMAGES[@]}"; do
    REPOSITORY=$(echo "$IMAGE" | cut -d'/' -f1)
    IMAGE_NAME=$(echo "$IMAGE" | cut -d'/' -f2-)

    sudo nerdctl pull "registry.hub.docker.com/$IMAGE"
    sudo nerdctl image tag "registry.hub.docker.com/$IMAGE" "$REGISTRY/$IMAGE_NAME"
    sudo nerdctl --insecure-registry push --snapshotter soci "$REGISTRY/$IMAGE_NAME"

    if [ "$IMAGE" = "clelange/cms-higgs-4l-full:latest" ] || [ "$IMAGE" = "rootproject/root:6.32.02-ubuntu24.04" ]|| [ "$IMAGE" = "library/centos:7" ]; then
        if [ ! -f /tmp/0431F9FA-6202-E311-8B98-002481E1501E.root ]; then
            wget http://opendata.cern.ch/record/9538/files/assets/cms/MonteCarlo2012/Summer12_DR53X/TTGJets_8TeV-madgraph/AODSIM/PU_RD1_START53_V7N-v1/10000/0431F9FA-6202-E311-8B98-002481E1501E.root -P /tmp
        fi
        sudo nerdctl pull "registry.cern.ch/snapshotter-benchmark/$IMAGE_NAME-esgz"
        sudo nerdctl image tag "registry.cern.ch/snapshotter-benchmark/$IMAGE_NAME-esgz" "$REGISTRY/$IMAGE_NAME-esgz"
        sudo nerdctl push --insecure-registry "$REGISTRY/$IMAGE_NAME-esgz"
    else
        sudo nerdctl pull "ghcr.io/stargz-containers/$IMAGE_NAME-esgz"
        sudo nerdctl image tag "ghcr.io/stargz-containers/$IMAGE_NAME-esgz" "$REGISTRY/$IMAGE_NAME-esgz"
        sudo nerdctl --insecure-registry push "$REGISTRY/$IMAGE_NAME-esgz"
        sudo nerdctl image rm "ghcr.io/stargz-containers/$IMAGE_NAME-esgz"
    fi

    sudo nerdctl image rm "$REGISTRY/$IMAGE_NAME-esgz"
    sudo nerdctl image rm "registry.hub.docker.com/$IMAGE"
    sudo nerdctl image rm "$REGISTRY/$IMAGE_NAME"
done
