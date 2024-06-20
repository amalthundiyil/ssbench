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

package framework

import (
	"context"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/containerd/containerd/remotes/docker/config"
)

func (proc *ContainerdProcess) PullImageFromRegistry(
	ctx context.Context,
	imageRef string,
	platform string) (containerd.Image, error) {
	opts := GetRemoteOpts(ctx, platform)
	opts = append(opts, containerd.WithResolver(GetResolver(ctx, imageRef)))
	image, pullErr := proc.Client.Pull(ctx, imageRef, opts...)
	if pullErr != nil {
		return nil, pullErr
	}
	return image, nil
}

func GetResolver(ctx context.Context, imageRef string) remotes.Resolver {
	var options = docker.ResolverOptions{
		Hosts: config.ConfigureHosts(context.TODO(), config.HostOptions{
			DefaultScheme: "http",
		}),
		Tracker: docker.NewInMemoryTracker(),
	}
	return docker.NewResolver(options)
}
