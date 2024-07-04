#!/bin/bash

set -xe

sudo systemctl stop cvmfs-snapshotter
sudo systemctl stop containerd

sudo cvmfs_config wipecache && sudo cvmfs_config killall

sudo systemctl start containerd
sudo systemctl start cvmfs-snapshotter

sudo nerdctl system prune -af