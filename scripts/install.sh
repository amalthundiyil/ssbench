#!/bin/bash

set -xe

if [ "x$PROJECT_DIR" = "x" ]; then 
    PROJECT_DIR="/home/vagrant/ssbench"
fi

mkdir -p $PROJECT_DIR/bin

# install other requirements
sudo apt install git make fuse -y

# install cvmfs
sudo apt install lsb-release -y
if [ ! -f /tmp/cvmfs-release-latest_all.deb ]; then
    wget https://ecsft.cern.ch/dist/cvmfs/cvmfs-release/cvmfs-release-latest_all.deb -O /tmp/cvmfs-release-latest_all.deb
fi
sudo dpkg -i /tmp/cvmfs-release-latest_all.deb
sudo apt update
sudo apt install -y cvmfs
sudo cvmfs_config setup
sudo sh -c "echo "CVMFS_HTTP_PROXY=DIRECT" > /etc/cvmfs/default.local"
sudo sh -c "echo "CVMFS_DEBUGLOG=/tmp/cvmfs.log" >> /etc/cvmfs/default.local"
sudo cvmfs_config reload

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
sudo tar Cxzvvf /usr/local /tmp/nerdctl-full-1.7.6-linux-amd64.tar.gz
sudo mkdir -p /etc/containerd
sudo cp $PROJECT_DIR/containerd_config.toml /etc/containerd/config.toml
sudo systemctl enable --now containerd

# soci

## soci build deps
sudo apt install zlib1g-dev gcc fuse unzip docker.io -y
wget -c https://github.com/google/flatbuffers/releases/download/v23.3.3/Linux.flatc.binary.g++-10.zip -P /tmp
sudo unzip -o /tmp/Linux.flatc.binary.g++-10.zip -d /usr/local

if [ ! -d /tmp/soci-snapshotter ]; then
    git clone https://github.com/awslabs/soci-snapshotter /tmp/soci-snapshotter
fi
cd /tmp/soci-snapshotter
git checkout 2f82461d214d2bc30843af32c52b7883b304db60
### Instead of sudo make
cd cmd/ && GO111MODULE=auto sudo /usr/local/go/bin/go build -o /tmp/soci-snapshotter/out/soci-snapshotter-grpc  -ldflags '-X github.com/awslabs/soci-snapshotter/version.Version=2f82461 -X github.com/awslabs/soci-snapshotter/version.Revision=2f82461d214d2bc30843af32c52b7883b304db60  -s -w '  ./soci-snapshotter-grpc
GO111MODULE=auto sudo /usr/local/go/bin/go build -o /tmp/soci-snapshotter/out/soci  -ldflags '-X github.com/awslabs/soci-snapshotter/version.Version=2f82461 -X github.com/awslabs/soci-snapshotter/version.Revision=2f82461d214d2bc30843af32c52b7883b304db60  -s -w '  ./soci
sudo cp /tmp/soci-snapshotter/out/* /usr/local/bin
sudo cp /tmp/soci-snapshotter/out/* $PROJECT_DIR/bin
# if [ ! -f /tmp/soci-snapshotter-0.6.1-linux-amd64.tar.gz ]; then
#     wget https://github.com/awslabs/soci-snapshotter/releases/download/v0.6.1/soci-snapshotter-0.6.1-linux-amd64.tar.gz -P /tmp
# fi 
# sudo tar -C /usr/local/bin -xvf /tmp/soci-snapshotter-0.6.1-linux-amd64.tar.gz soci soci-snapshotter-grpc
# sudo tar -C $PROJECT_DIR/bin -xvf /tmp/soci-snapshotter-0.6.1-linux-amd64.tar.gz soci soci-snapshotter-grpc
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
/usr/local/go/bin/go build -o out/cvmfs_snapshotter -ldflags '-X main.Version=2.11'                                                                                              
cp /tmp/cvmfs/snapshotter/out/cvmfs_snapshotter $PROJECT_DIR/bin/cvmfs_snapshotter
cp /tmp/cvmfs/snapshotter/out/cvmfs_snapshotter /usr/local/bin/cvmfs_snapshotter
wget https://raw.githubusercontent.com/cvmfs/cvmfs/42e04529dc8eccb52bf62b27b220aa54b660681a/snapshotter/script/config/etc/systemd/system/cvmfs-snapshotter.service -O /usr/local/lib/systemd/system/cvmfs-snapshotter.service
mkdir -p /etc/containerd-cvmfs-grpc && touch /etc/containerd-cvmfs-grpc/config.toml

cd $PROJECT_DIR 

sudo systemctl daemon-reload
sudo systemctl start soci-snapshotter
sudo systemctl start stargz-snapshotter
sudo systemctl start cvmfs-snapshotter

/usr/local/go/bin/go build -o $PROJECT_DIR/bin/ssbench $PROJECT_DIR/cmd/main.go

