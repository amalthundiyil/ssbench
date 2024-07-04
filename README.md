# ssbench

ssbench (SnapShotter BENCHmark) is a benchmarking tool for [stargz](https://github.com/containerd/stargz-snapshotter), [soci](https://github.com/awslabs/soci-snapshotter) and [cvmfs](https://github.com/cvmfs/cvmfs/tree/devel/snapshotter) containerd snapshotter implementations.

## Usage

```
go build -o bin/ssbench ./cmd/main.go
cd bin
sudo ./ssbench
```

## Thanks

This project was made possible thanks to previous work done by:
- AWS Labs [soci-snapshotter benchmarker](https://github.com/awslabs/soci-snapshotter/tree/main/benchmark)
- Lange et al [Frontiers in Big Data '21](https://www.frontiersin.org/articles/10.3389/fdata.2021.673163/full)
