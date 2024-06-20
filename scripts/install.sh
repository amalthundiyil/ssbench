#!/bin/bash

set -xe

sudo apt install git make fuse -y

PROJECT_DIR="/home/ubuntu/ssbench"
# PROJECT_DIR="/home/amal/work/ssbench"

mkdir -p $PROJECT_DIR/bin

# go
if [ ! -f /tmp/go1.22.0.linux-amd64.tar.gz ]; then
    wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz -P /tmp
fi
tar -C /usr/local -xvf /tmp/go1.22.0.linux-amd64.tar.gz

echo "export GOPATH=$HOME/go" >> ~/.bashrc
echo "export PATH=\$PATH:/usr/local/go/bin:\$GOPATH/bin" >> ~/.bashrc
source ~/.bashrc

# nerdctl
if [ ! -f /tmp/nerdctl-full-1.7.6-linux-amd64.tar.gz ]; then
    wget https://github.com/containerd/nerdctl/releases/download/v1.7.6/nerdctl-full-1.7.6-linux-amd64.tar.gz -P /tmp
fi
tar Cxzvvf /usr/local /tmp/nerdctl-full-1.7.6-linux-amd64.tar.gz
sudo systemctl enable --now containerd

# soci
if [ ! -f /tmp/soci-snapshotter-0.6.1-linux-amd64.tar.gz ]; then
    wget https://github.com/awslabs/soci-snapshotter/releases/download/v0.6.1/soci-snapshotter-0.6.1-linux-amd64.tar.gz -P /tmp
fi 
sudo tar -C /usr/local/bin -xvf /tmp/soci-snapshotter-0.6.1-linux-amd64.tar.gz soci soci-snapshotter-grpc
sudo tar -C $PROJECT_DIR/bin -xvf /tmp/soci-snapshotter-0.6.1-linux-amd64.tar.gz soci soci-snapshotter-grpc
wget https://raw.githubusercontent.com/awslabs/soci-snapshotter/main/soci-snapshotter.service -O /usr/local/lib/systemd/system/soci-snapshotter.service

# stargz
if [ ! -f /tmp/stargz-snapshotter-v0.15.1-linux-amd64.tar.gz ]; then
    wget https://github.com/containerd/stargz-snapshotter/releases/download/v0.15.1/stargz-snapshotter-v0.15.1-linux-amd64.tar.gz -P /tmp
fi
sudo tar -C /usr/local/bin -xvf /tmp/stargz-snapshotter-v0.15.1-linux-amd64.tar.gz
sudo tar -C $PROJECT_DIR/bin -xvf /tmp/stargz-snapshotter-v0.15.1-linux-amd64.tar.gz

# cvmfs
if [ ! -d /tmp/cvmfs ]; then
    git clone https://github.com/cvmfs/cvmfs /tmp/cvmfs
fi
cd /tmp/cvmfs/snapshotter
make
cp /tmp/cvmfs/snapshotter/out/cvmfs_snapshotter $PROJECT_DIR/bin/cvmfs_snapshotter
cp /tmp/cvmfs/snapshotter/out/cvmfs_snapshotter /usr/local/bin/cvmfs_snapshotter
wget https://raw.githubusercontent.com/cvmfs/cvmfs/42e04529dc8eccb52bf62b27b220aa54b660681a/snapshotter/script/config/etc/systemd/system/cvmfs-snapshotter.service -O /usr/local/lib/systemd/system/cvmfs-snapshotter.service
mkdir -p /etc/containerd-cvmfs-grpc && touch /etc/containerd-cvmfs-grpc/config.toml


cd $PROJECT_DIR 

sudo systemctl daemon-reload
sudo systemctl start soci-snapshotter
sudo systemctl start stargz-snapshotter
sudo systemctl start cvmfs-snapshotter

go build -o $PROJECT_DIR/bin/ssbench $PROJECT_DIR/cmd/main.go

