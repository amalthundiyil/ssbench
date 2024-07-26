package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
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
	imageRef := "docker.io/library/redis:alpine"
	imageName := strings.Split(strings.Split(imageRef, "/")[2], ":")[0]

	image, err := client.Pull(ctx, imageRef,
		containerd.WithPullUnpack,
		containerd.WithPullSnapshotter("stargz"),
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

	containerdId := fmt.Sprintf("%s-%d", imageName, time.Now().UnixNano())
	rootfs, err := filepath.Abs(containerdId)
	if err != nil {
		log.Fatalf("Error rootfs: %s\n", err)
	}
	if err := os.MkdirAll(rootfs, 0770); err != nil {
		log.Fatalf("Error making: %s\n", err)
	}
	cleanupFolder := func() {
		if err = os.RemoveAll(rootfs); err != nil {
			log.Fatalf("Error delete the folder: %s", err)
		}
	}
	defer cleanupFolder()

	container, err := client.NewContainer(
		ctx,
		containerdId,
		containerd.WithNewSpec(oci.WithImageConfig(image)),
		containerd.WithImage(image),
		containerd.WithSnapshotter("stargz"),
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

	// command := "echo '#include <stdio.h>\nint main() { printf(\"Hello World\\n\"); return 0; }' > /tmp/main.c && gcc -o /tmp/a.out /tmp/main.c && /tmp/a.out"
	// cioCreator := cio.BinaryIO("/usr/bin/sh", map[string]string{"-c": command})
	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		log.Fatalf("Error creating task: %s\n", err)
	}
	defer task.Delete(ctx)
}
