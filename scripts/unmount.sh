#!/bin/bash

BASE_PATTERN="/var/lib/containerd-*-grpc/snapshotter/snapshots"
base_dirs=$(ls -d /var/lib/containerd-*-grpc/snapshotter/snapshots 2>/dev/null)

if [ -z "$base_dirs" ]; then
    echo "No directories matching the pattern found."
    exit 1
fi

for base_dir in $base_dirs; do
    echo "Searching for FUSE filesystems in $base_dir..."

    mount_points=$(mount | grep "type fuse.rawBridge" | awk '{print $3}' | grep "^$base_dir")

    for mount_point in $mount_points; do
        echo "Unmounting $mount_point..."
        sudo umount "$mount_point"
        if [ $? -eq 0 ]; then
            echo "Successfully unmounted $mount_point"
        else
            echo "Failed to unmount $mount_point"
        fi
    done
done

remaining_mounts=$(mount | grep "type fuse.rawBridge")
if [ -z "$remaining_mounts" ]; then
    echo "All FUSE filesystems have been successfully unmounted."
else
    echo "Some FUSE filesystems are still mounted:"
    echo "$remaining_mounts"
fi
