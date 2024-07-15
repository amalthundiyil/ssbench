/*
   Copyright The Soci Snapshotter Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package benchmark

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/contrib/nvidia"
	"github.com/containerd/containerd/mount"
	"github.com/containerd/containerd/oci"
	runtimespec "github.com/opencontainers/runtime-spec/specs-go"
)

const testContainerID = "TEST_RUN_CONTAINER"

type ImageDescriptor struct {
	ShortName       string       `json:"short_name"`
	ImageRef        string       `json:"image_ref"`
	SociIndexDigest string       `json:"soci_index_digest"`
	ReadyLine       string       `json:"ready_line"`
	Command         string       `json:"command"`
	TimeoutSec      int64        `json:"timeout_sec"`
	ImageOptions    ImageOptions `json:"options"`
	Mount           mount.Mount  `json:"mount"`
}

func (i *ImageDescriptor) Timeout() time.Duration {
	if i.TimeoutSec <= 0 {
		return 180 * time.Second
	}
	return time.Duration(i.TimeoutSec) * time.Second
}

// ImageOptions contains image-specific options needed to run the tests
type ImageOptions struct {
	// Net indicicates the container's network mode. If set to "host" then the container will have host networking, otherwise no networking.
	Net string `json:"net"`
	// Mounts are any mounts needed by the container
	Mounts []runtimespec.Mount `json:"mounts"`
	// Gpu is whether the container needs GPUs. If true, all GPUs are mounted in the container
	Gpu bool `json:"gpu"`
	// Env is any environment variables needed by the containerd
	Env []string `json:"env"`
	// ShmSize is the size of /dev/shm to be used inside the container
	ShmSize int64 `json:"shm_size"`
}

// ContainerOpts creates a set of NewContainerOpts from an ImageDescriptor and a containerd.Image
// The options can be used directly when launching a container
func (i *ImageDescriptor) ContainerOpts(image containerd.Image, o ...containerd.NewContainerOpts) []containerd.NewContainerOpts {
	var opts []containerd.NewContainerOpts
	var ociOpts []oci.SpecOpts

	opts = append(opts, o...)
	id := fmt.Sprintf("%s-%d", testContainerID, time.Now().UnixNano())
	opts = append(opts, containerd.WithNewSnapshot(id, image))
	ociOpts = append(ociOpts, oci.WithImageConfig(image))
	if len(i.ImageOptions.Mounts) > 0 {
		ociOpts = append(ociOpts, oci.WithMounts(i.ImageOptions.Mounts))
	}
	if i.ImageOptions.Gpu {
		ociOpts = append(ociOpts, nvidia.WithGPUs(nvidia.WithAllDevices, nvidia.WithAllCapabilities))
	}
	if len(i.ImageOptions.Env) > 0 {
		ociOpts = append(ociOpts, oci.WithEnv(i.ImageOptions.Env))
	}
	if i.ImageOptions.ShmSize > 0 {
		ociOpts = append(ociOpts, oci.WithDevShmSize(i.ImageOptions.ShmSize))
	}
	if i.ImageOptions.Net == "host" {
		hostname, err := os.Hostname()
		if err != nil {
			panic(fmt.Errorf("get hostname: %w", err))
		}
		ociOpts = append(ociOpts,
			oci.WithHostNamespace(runtimespec.NetworkNamespace),
			oci.WithHostHostsFile,
			oci.WithHostResolvconf,
			oci.WithEnv([]string{fmt.Sprintf("HOSTNAME=%s", hostname)}),
		)
	}

	opts = append(opts, containerd.WithNewSpec(ociOpts...))
	return opts
}

func GetImageList(file string) ([]ImageDescriptor, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return GetImageListFromJSON(f)

}

func GetImageListFromJSON(r io.Reader) ([]ImageDescriptor, error) {
	var images []ImageDescriptor
	err := json.NewDecoder(r).Decode(&images)
	if err != nil {
		return nil, err
	}
	return images, nil
}

func GetCommitHash() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func GetDefaultWorkloads() []ImageDescriptor {
	return []ImageDescriptor{
		{
			ShortName:       "Python3.9",
			ImageRef:        "localhost:5000/python:3.9",
			SociIndexDigest: "sha256:05d4b7931cbef9792f3058e51b40f5b0b16bfdd9e5ebdc4857a1ffa135329865",
			ReadyLine:       "Hello World",
			Command:         "python3 -c \"print('Hello World')\"",
		},
		{
			ShortName:       "Gcc11.2.0",
			ImageRef:        "localhost:5000/gcc:11.2.0",
			SociIndexDigest: "sha256:3e8e261a6ab60fb93afaf75b3c5cc4455fa147ef7e779eb8509d18bf051fd0e0",
			ReadyLine:       "Hello World",
			Command:         "echo '#include <stdio.h>\nint main() { printf(\"Hello World\\n\"); return 0; }' > /tmp/main.c && gcc -o /tmp/a.out /tmp/main.c && /tmp/a.out",
		},
		{
			ShortName:       "Cms-higgs-4l-full",
			ImageRef:        "localhost:5000/cms-higgs-4l-full:latest",
			SociIndexDigest: "sha256:7061ecf4a6b00812dad0fe9b0d49f7c7db59974d23912e5f448de5990b42b0f0",
			ReadyLine:       "Report end",
			Command:         "export CMS_INPUT_FILES=file:///tmp/0431F9FA-6202-E311-8B98-002481E1501E.root && /opt/cms/entrypoint.sh cmsRun /configs/demoanalyzer_cfg_level4MC.py",
			Mount:           mount.Mount{Type: "bind", Source: "/tmp/0431F9FA-6202-E311-8B98-002481E1501E.root", Target: "/tmp/0431F9FA-6202-E311-8B98-002481E1501E.root"},
		},
	}
}

func WriteDefaultConfig() {
	tomlContent := `
version = 2

[plugins."io.containerd.grpc.v1.cri".containerd]
    disable_snapshot_annotations = false
  
[proxy_plugins]
    [proxy_plugins.cvmfs-snapshotter]
        type = "snapshot"
        address = "/tmp/containerd-cvmfs-grpc/containerd-cvmfs-grpc.sock"
    [proxy_plugins.stargz]
        type = "snapshot"
        address = "/tmp/containerd-stargz-grpc/containerd-stargz-grpc.sock"
    [proxy_plugins.soci]
        type = "snapshot"
        address = "/tmp/containerd-soci-grpc/containerd-soci-grpc.sock"
`

	filePath := "/tmp/containerd_config.toml"

	if _, err := os.Stat(filePath); err == nil {
		log.Printf("File %s already exists and will be overwritten", filePath)
	} else if !os.IsNotExist(err) {
		log.Fatalf("Error checking if file exists: %v", err)
	}

	err := os.WriteFile(filePath, []byte(tomlContent), 0644)
	if err != nil {
		log.Fatalf("Error writing to file: %v", err)
	}
}
