#!/bin/bash

set -xe

sudo systemctl stop stargz-snapshotter
sudo systemctl stop soci-snapshotter
sudo systemctl stop cvmfs-snapshotter
sudo systemctl stop containerd

sudo cvmfs_config wipecache
sudo cvmfs_config killall

sudo rm -rf /var/lib/containerd-stargz-grpc
sudo rm -rf /var/lib/containerd-cvmfs-grpc
sudo rm -rf /var/lib/containerd-soci-grpc

sudo mkdir -p /var/lib/containerd-stargz-grpc/snapshotter/snapshots
sudo mkdir -p /var/lib/containerd-cvmfs-grpc/snapshotter/snapshots
sudo mkdir -p /var/lib/containerd-soci-grpc/snapshotter/snapshots

sudo systemctl start containerd
sudo systemctl start stargz-snapshotter
sudo systemctl start soci-snapshotter
sudo systemctl start cvmfs-snapshotter
