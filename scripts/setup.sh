#!/bin/bash

set -xe

IMAGES=(
    "library/python:3.9" 
    "library/gcc:11.2.0"
    "clelange/cms-higgs-4l-full:latest"
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

    if [ "$IMAGE" = "clelange/cms-higgs-4l-full:latest" ]; then
        echo "Optimizing image"
        # if [ ! -f /tmp/0431F9FA-6202-E311-8B98-002481E1501E.root ]; then
        #     wget http://opendata.cern.ch/record/9538/files/assets/cms/MonteCarlo2012/Summer12_DR53X/TTGJets_8TeV-madgraph/AODSIM/PU_RD1_START53_V7N-v1/10000/0431F9FA-6202-E311-8B98-002481E1501E.root -P /tmp
        # fi
        # sudo ctr-remote image optimize --oci --mount=type=bind,source=/tmp/0431F9FA-6202-E311-8B98-002481E1501E.root,destination=/tmp/0431F9FA-6202-E311-8B98-002481E1501E.root,options=bind:ro --entrypoint='["/bin/bash", "-c"]' --args='["export CMS_INPUT_FILES=file:///tmp/0431F9FA-6202-E311-8B98-002481E1501E.root && /opt/cms/entrypoint.sh cmsRun /configs/demoanalyzer_cfg_level4MC.py"]' --period 100 "registry.hub.docker.com/clelange/cms-higgs-4l-full:latest" "$REGISTRY/$IMAGE_NAME-esgz"
        # sudo nerdctl push --insecure-registry "$REGISTRY/$IMAGE_NAME-esgz"
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
