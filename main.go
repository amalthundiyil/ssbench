package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"syscall"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/containerd/containerd/pkg/snapshotters"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	client, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		return err
	}
	defer client.Close()

	ctx := namespaces.WithNamespace(context.Background(), "default")

	imageRef := "docker.io/library/python:3.9"
	command := "python3 -c \"print('Hello World')\""
	// imageRef := "docker.io/library/gcc:11.2.0"
	// command := "echo '#include <stdio.h>\nint main() { printf(\"Hello World\\n\"); return 0; }' > /tmp/main.c && gcc -o /tmp/a.out /tmp/main.c && /tmp/a.out"
	imageName := strings.Split(strings.Split(imageRef, "/")[2], ":")[0]
	log.Default().Println("Pulling image")
	image, err := client.Pull(ctx, imageRef,
		containerd.WithPullUnpack,
		containerd.WithPullSnapshotter("cvmfs-snapshotter"),
		containerd.WithImageHandlerWrapper(snapshotters.AppendInfoHandlerWrapper(imageRef)),
	)
	if err != nil {
		return err
	}
	log.Default().Println("Done Pulling")

	cleanupImage := func() {
		err = client.ImageService().Delete(ctx, image.Name())
		if err != nil {
			fmt.Println(err)
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
		return err
	}
	defer container.Delete(ctx, containerd.WithSnapshotCleanup)

	task, err := container.NewTask(ctx, cio.BinaryIO("/usr/bin/sh", map[string]string{"-c": command}))
	if err != nil {
		return err
	}
	defer task.Delete(ctx)

	exitStatusC, err := task.Wait(ctx)
	if err != nil {
		fmt.Println(err)
	}
	if err := task.Start(ctx); err != nil {
		return err
	}

	if err := task.Kill(ctx, syscall.SIGTERM); err != nil {
		return err
	}

	status := <-exitStatusC
	code, _, err := status.Result()
	if err != nil {
		return err
	}
	fmt.Printf(imageName+" exited with status: %d\n", code)

	return nil
}
