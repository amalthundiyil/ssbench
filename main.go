package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/containerd/containerd/pkg/snapshotters"
	"github.com/opencontainers/runtime-spec/specs-go"
)

func main() {
	client, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		log.Fatalf("Error creating client: %s\n", err)
	}
	defer client.Close()

	ctx := namespaces.WithNamespace(context.Background(), "default")
	imageRef := "docker.io/clelange/cms-higgs-4l-full:latest"
	imageName := strings.Split(strings.Split(imageRef, "/")[2], ":")[0]
	snapshotter := "cvmfs-snapshotter"

	image, err := client.Pull(ctx, imageRef,
		containerd.WithSchema1Conversion,
		containerd.WithPullUnpack,
		containerd.WithPullSnapshotter(snapshotter),
		containerd.WithImageHandlerWrapper(snapshotters.AppendInfoHandlerWrapper(imageRef)),
	)
	if err != nil {
		log.Fatalf("Error pulling image: %s\n", err)
	}

	cleanupImage := func() {
		err = client.ImageService().Delete(ctx, image.Name())
		if err != nil {
			log.Printf("Error deleting image: %s\n", err)
		}
	}
	defer cleanupImage()

	command := "export CMS_INPUT_FILES=file:///tmp/0431F9FA-6202-E311-8B98-002481E1501E.root && /opt/cms/entrypoint.sh cmsRun /configs/demoanalyzer_cfg_level4MC.py"
	mounts := []specs.Mount{
		{
			Type:        "bind",
			Source:      "/tmp/0431F9FA-6202-E311-8B98-002481E1501E.root",
			Destination: "/tmp/0431F9FA-6202-E311-8B98-002481E1501E.root",
			Options:     []string{"rbind", "rw"},
		},
	}

	containerdId := fmt.Sprintf("%s-%d", imageName, time.Now().UnixNano())
	container, err := client.NewContainer(
		ctx,
		containerdId,
		containerd.WithSnapshotter(snapshotter),
		containerd.WithNewSnapshot(containerdId+"-snapshot", image),
		containerd.WithNewSpec(
			oci.WithImageConfig(image),
			oci.WithProcessArgs("/bin/sh", "-c", command),
			oci.WithMounts(mounts)),
	)
	if err != nil {
		log.Fatalf("Error creating new container: %s\n", err)
	}

	cleanupContainer := func() {
		err = container.Delete(ctx, containerd.WithSnapshotCleanup)
		if err != nil {
			log.Printf("Error deleting container: %s\n", err)
		}
	}
	defer cleanupContainer()

	r, w := io.Pipe()
	cioCreator := cio.NewCreator(cio.WithStreams(nil, w, w))
	task, err := container.NewTask(ctx, cioCreator)
	if err != nil {
		log.Fatalf("Error creating task: %s\n", err)
	}
	defer task.Delete(ctx)

	go func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			log.Printf("Error reading task output: %s\n", err)
		}
	}()

	statusC, err := task.Wait(ctx)
	if err != nil {
		log.Fatalf("Error waiting for task: %s\n", err)
	}

	if err := task.Start(ctx); err != nil {
		log.Fatalf("Error starting task: %s\n", err)
	}

	fmt.Println("Task started successfully.")

	status := <-statusC
	code, _, err := status.Result()
	if err != nil {
		log.Fatalf("Error getting task result: %s\n", err)
	}

	w.Close()

	fmt.Printf("Task exited with status code: %d\n", code)

	if code == 0 {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
