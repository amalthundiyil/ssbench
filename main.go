package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/containerd/containerd/pkg/snapshotters"
)

func main() {
	client, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		log.Fatalf("Error creating client: %s\n", err)
	}
	defer client.Close()

	ctx := namespaces.WithNamespace(context.Background(), "default")
	imageRef := "docker.io/library/python:3.9"
	imageName := strings.Split(strings.Split(imageRef, "/")[2], ":")[0]

	image, err := client.Pull(ctx, imageRef,
		containerd.WithPullUnpack,
		containerd.WithPullSnapshotter("cvmfs-snapshotter"),
		containerd.WithImageHandlerWrapper(snapshotters.AppendInfoHandlerWrapper(imageRef)),
	)
	if err != nil {
		log.Fatalf("Error pulling image: %s\n", err)
	}

	cleanupImage := func() {
		err = client.ImageService().Delete(ctx, image.Name())
		if err != nil {
			log.Fatalf("Error deleting image: %s\n", err)
		}
	}
	defer cleanupImage()

	container, err := client.NewContainer(
		ctx,
		fmt.Sprintf("%s-%d", imageName, time.Now().UnixNano()),
		containerd.WithImage(image),
		containerd.WithNewSnapshot(fmt.Sprintf("%s-%d-snapshot", imageName, time.Now().UnixNano()), image),
		containerd.WithNewSpec(oci.WithImageConfig(image)),
		containerd.WithSnapshotter("cvmfs-snapshotter"),
	)
	if err != nil {
		log.Fatalf("Error creating new container: %s\n", err)
	}
	cleanupContainer := func() {
		err = container.Delete(ctx, containerd.WithSnapshotCleanup)
		if err != nil {
			log.Fatalf("Error deleting container: %s\n", err)
		}
	}
	defer cleanupContainer()
}
