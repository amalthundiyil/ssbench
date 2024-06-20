package benchmark

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/amalthundiyil/ssbench/pkg/framework"
	"github.com/containerd/containerd"
	ctdsnapshotters "github.com/containerd/containerd/pkg/snapshotters"
)

type CvmfsProcess struct {
	command *exec.Cmd
	address string
	root    string
	stdout  *os.File
	stderr  *os.File
}

type CvmfsContainerdProcess struct {
	*framework.ContainerdProcess
}

func StartCvmfs(
	cvmfsBinary string,
	cvmfsAddress string,
	cvmfsConfig string,
	cvmfsRoot string,
	outputDir string) (*CvmfsProcess, error) {
	cvmfsCmd := exec.Command(cvmfsBinary,
		"-address", cvmfsAddress,
		"-config", cvmfsConfig,
		"-log-level", "debug",
		"-root", cvmfsRoot)
	err := os.MkdirAll(outputDir, 0777)
	if err != nil {
		return nil, err
	}
	stdoutFile, err := os.Create(outputDir + "/cvmfs-snapshotter-stdout")
	if err != nil {
		return nil, err
	}
	cvmfsCmd.Stdout = stdoutFile
	stderrFile, err := os.Create(outputDir + "/cvmfs-snapshotter-stderr")
	if err != nil {
		return nil, err
	}
	cvmfsCmd.Stderr = stderrFile
	err = cvmfsCmd.Start()
	if err != nil {
		fmt.Printf("Cvmfs process failed to start %v\n", err)
		return nil, err
	}

	// The cvmfs-snapshotter-grpc is not ready to be used until the
	// unix socket file is created
	sleepCount := 0
	loopExit := false
	for !loopExit {
		time.Sleep(1 * time.Second)
		sleepCount++
		if _, err := os.Stat(cvmfsAddress); err == nil {
			loopExit = true
		}
		if sleepCount > 15 {
			return nil, errors.New("could not create .sock in time")
		}
	}
	return &CvmfsProcess{
		command: cvmfsCmd,
		address: cvmfsAddress,
		root:    cvmfsRoot,
		stdout:  stdoutFile,
		stderr:  stderrFile}, nil
}

func (proc *CvmfsProcess) StopProcess() {
	if proc.stdout != nil {
		proc.stdout.Close()
	}
	if proc.stderr != nil {
		proc.stderr.Close()
	}
	if proc.command != nil {
		proc.command.Process.Kill()
	}
	err := os.RemoveAll(proc.address)
	if err != nil {
		fmt.Printf("Error removing cvmfs process address: %v\n", err)
	}

	snapshotDir := proc.root + "/snapshotter/snapshots/"
	snapshots, err := os.ReadDir(snapshotDir)
	if err != nil {
		fmt.Printf("Could not read dir: %s\n", snapshotDir)
	}

	for _, s := range snapshots {
		mountpoint := snapshotDir + s.Name() + "/fs"
		_ = syscall.Unmount(mountpoint, syscall.MNT_FORCE)
	}
	err = os.RemoveAll(proc.root)
	if err != nil {
		fmt.Printf("Error removing cvmfs process root: %v\n", err)
	}
}

func (proc *CvmfsContainerdProcess) CvmfsPullImageFromRegistry(
	ctx context.Context,
	imageRef string) (containerd.Image, error) {
	image, err := proc.Client.Pull(ctx, imageRef, []containerd.RemoteOpt{
		containerd.WithResolver(framework.GetResolver(ctx, imageRef)),
		//nolint:staticcheck
		containerd.WithSchema1Conversion, //lint:ignore SA1019
		containerd.WithPullUnpack,
		containerd.WithPullSnapshotter("cvmfs-snapshotter"),
		containerd.WithImageHandlerWrapper(ctdsnapshotters.AppendInfoHandlerWrapper(imageRef)),
	}...)
	if err != nil {
		fmt.Printf("Cvmfs pull failed: %v\n", err)
		return nil, err
	}
	return image, nil
}
