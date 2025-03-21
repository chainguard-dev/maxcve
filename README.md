# _MAXIMUM CVEs_

[![Release](https://github.com/chainguard-dev/maxcve/actions/workflows/release.yaml/badge.svg)](https://github.com/chainguard-dev/maxcve/actions/workflows/release.yaml)

This repo generates a container image that maximizes the number of CVEs in the image, while minimizing the size of the image.

```
$ grype ghcr.io/chainguard-dev/maxcve/maxcve 1> /dev/null
   ├── ✔ Packages                        [48,215 packages]
   └── ✔ Executables                     [0 executables]
 ✔ Scanned for vulnerabilities     [290565 vulnerability matches]
   ├── by severity: 5968 critical, 50545 high, 38097 medium, 1390 low, 0 negligible (194565 unknown)
   └── by status:   282221 fixed, 8344 not-fixed, 0 ignored
```

(As of March 28, 2024)

Or, if you prefer to consume data visually:

![](severity.png)

_Zero negligible vulns, nice!_

![](installed.png)

_Real minimal base image for scale_

### Development

```
go run . ttl.sh/maxcve
```

### How it works

To minimize size, the image doesn't actually contain any packages. In fact, it only contains two files:

1. `/etc/os-release`, which tells scanners the image is a [Wolfi](https://wolfi.dev) image.
1. `/lib/apk/db/installed`, which tells scanners what packages the image contains -- i.e., that it contains every version of every package that Wolfi has ever produced.

Wolfi aims to reduce the number of vulnerable packages by producing new fixed packages as soon as possible. But, along the way, it also produces lots and _lots_ of packages, and those packages over time _do_ have vulnerabilities discovered in them. This image claims to contain all of them.

Amusingly, it takes about 500ms to build and push the image, and almost two minutes to scan it.

### Why?

Aside from being fun, this image demonstrates how scanners work -- and importantly, how they _don't_ work.

At their most basic, scanners require images (1) tell them what OS they are, and (2) tell them what packages they contain. This image does both, but it does so in a way that is misleading.

For a similar (but opposite) demonstration of this, see [Malicious Compliance: Reflections on Trusting Container Scanners](https://www.youtube.com/watch?v=9weGi0csBZM). In that talk, they mislead the scanner into finding fewer CVEs in the presence of vulnerable packages. In this demonstration, we mislead the scanner into finding vulnerabilities without installing any packages.

### Proof

The following script demonstrates that all the CVEs in the image are entirely due to the existence of the `/lib/apk/db/installed` file, which lists all the packages that are "installed".

Running a Grype scan after removing that file from the image results in a Grype scan with zero CVEs:

```
grype ghcr.io/chainguard-dev/maxcve/maxcve 1> /dev/null
TEMP_DIR=$(mktemp -d) && \
	crane export ghcr.io/chainguard-dev/maxcve/maxcve:latest - | tar -xvf - -C "$TEMP_DIR" && \
	rm -f "$TEMP_DIR/lib/apk/db/installed" && \
	tar -C "$TEMP_DIR" -cf - . | docker import - maxcve:noapkdb && \
	rm -rf "$TEMP_DIR"
grype maxcve:noapkdb 1> /dev/null
```


