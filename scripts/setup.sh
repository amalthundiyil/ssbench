#!/bin/bash

set -xe

IMAGES=(
    "registry.hub.docker.com/library/python:3.9" 
    "registry.hub.docker.com/library/gcc:11.2.0"
    "registry.hub.docker.com/clelange/cms-higgs-4l-full:latest"
)

REGISTRY="localhost:5000"

# sudo soci index rm `sudo soci index ls -q`
if [[ `sudo nerdctl ps -aq` ]]; then
    sudo nerdctl rm -f `sudo nerdctl ps -aq`
fi

sudo nerdctl run -d -p 5000:5000 --restart=always --name registry registry:2.7

# pull images from remotes and then push to local
for IMAGE in ${IMAGES[@]}; do
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
done

if [ ! -f /tmp/0431F9FA-6202-E311-8B98-002481E1501E.root ]; then
    wget http://opendata.cern.ch/record/9538/files/assets/cms/MonteCarlo2012/Summer12_DR53X/TTGJets_8TeV-madgraph/AODSIM/PU_RD1_START53_V7N-v1/10000/0431F9FA-6202-E311-8B98-002481E1501E.root -P /tmp
fi