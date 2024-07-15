package main

import (
	"context"
	"fmt"
	"os"

	"github.com/awslabs/soci-snapshotter/fs/source"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/pkg/snapshotters"
	"github.com/containerd/containerd/remotes/docker"
)

func main() {
	var address = "/run/containerd/containerd.sock"
	var ref = "localhost:5000/python:3.9"
	var sociIndexDigest = "sha256:974c88873f74c89cc87c69904fae9ef9362e5aab9a555f1069e7814df2cfb681"

	client, err := containerd.New(address, containerd.WithDefaultNamespace("default"))
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer client.Close()

	options := docker.ResolverOptions{
		Hosts: docker.ConfigureDefaultRegistries(docker.WithPlainHTTP(func(string) (bool, error) { return true, nil })),
	}

	registryHosts, err := options.Hosts("localhost:5000")
	if err == nil {
		for _, registryHost := range registryHosts {
			fmt.Printf("%s %s \n", registryHost.Scheme, registryHost.Host)
		}
	}

	_, err = client.Pull(context.TODO(), ref,
		containerd.WithResolver(docker.NewResolver(options)),
		containerd.WithPullUnpack,
		containerd.WithImageHandlerWrapper(source.AppendDefaultLabelsHandlerWrapper(sociIndexDigest, snapshotters.AppendInfoHandlerWrapper(ref))))
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("Success")
}
