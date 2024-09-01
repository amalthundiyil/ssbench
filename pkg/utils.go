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
	ociOpts = append(ociOpts, oci.WithProcessArgs("/bin/sh", "-c", i.Command))
	if len(i.ImageOptions.Mounts) > 0 {
		ociOpts = append(ociOpts, oci.WithMounts(i.ImageOptions.Mounts))
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
	mounts := []runtimespec.Mount{
		{
			Type:        "bind",
			Source:      "/tmp/0431F9FA-6202-E311-8B98-002481E1501E.root",
			Destination: "/tmp/0431F9FA-6202-E311-8B98-002481E1501E.root",
			Options:     []string{"rbind", "rw"},
		},
	}

	return []ImageDescriptor{
		{
			ShortName:       "python:3.9",
			ImageRef:        "localhost:5000/python:3.9",
			SociIndexDigest: "sha256:fe7d8fc5c26736d806f1cc0407ec190f81f8485c6bad127bc3696bd1748f5420",
			ReadyLine:       "Hello World",
			Command:         "python3 -c \"print('Hello World')\"",
		},
		{
			ShortName:       "gcc:11.2.0",
			ImageRef:        "localhost:5000/gcc:11.2.0",
			SociIndexDigest: "sha256:7dc1d7344df1ab0d32b24f7590a1f08dc1ca06a2e62d63caf37c1566fc6f32f3",
			ReadyLine:       "Hello World",
			Command:         `echo '#include <stdio.h>\n int main() { printf("Hello World\\n"); return 0; }' > /tmp/main.c && gcc -o /tmp/a.out /tmp/main.c && /tmp/a.out`,
		},
		{
			ShortName:       "root:6.32.02-ubuntu24.04",
			ImageRef:        "localhost:5000/root:6.32.02-ubuntu24.04",
			SociIndexDigest: "sha256:a844051b4b477cc0362ef8fab2e5884185b9b091661f27f2626577094bf08be2",
			ReadyLine:       "Hello World",
			Command:         `echo '#include <stdio.h>\n int main() { printf("Hello World\\n"); return 0; }' > /tmp/main.c && root -l -q /tmp/main.c`,
		},
		{
			ShortName:       "cms-higgs-4l-full:latest",
			ImageRef:        "localhost:5000/cms-higgs-4l-full:latest",
			SociIndexDigest: "sha256:39bdb13731107f67e8d691bd8984a5e9191eaa9ca113e9cf48241d7cdabb864f",
			ReadyLine:       "Report end",
			Command:         "export CMS_INPUT_FILES=file:///tmp/0431F9FA-6202-E311-8B98-002481E1501E.root && /opt/cms/entrypoint.sh cmsRun /configs/demoanalyzer_cfg_level4MC.py",
			ImageOptions:    ImageOptions{Mounts: mounts},
		},
		{
			ShortName:       "centos:7",
			ImageRef:        "localhost:5000/centos:7",
			SociIndexDigest: "sha256:fab72f18e60eaafc48974dac821d3db87c2fec166b821bc2079bb072007fa333",
			ReadyLine:       "Hello World",
			Command:         "echo 'Hello World'",
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
