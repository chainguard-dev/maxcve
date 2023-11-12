# _MAXIMUM CVEs_

This repo generates a container image that maximizes the number of CVEs in the image, while minimizing the size of the image.

The result is a 1.8 MB image that reports as having more than 20,000 CVEs. That's roughly one CVE per 100 bytes!

```
$ time grype $(go run .) > /dev/null
2023/11/12 11:07:09 wrote /lib/apk/db/installed
2023/11/12 11:07:09 wrote /etc/os-release
2023/11/12 11:07:09 wrote ttl.sh/maxcve@sha256:c43609f71b0bf2d3f317d6347291bc070c09aab40cdcae5a16b723ea596620ab
 ✔ Vulnerability DB                [no update available]
 ✔ Loaded image                                                                ttl.sh/maxcve@sha256:c43609f71b0bf2d3f317d6347291bc070c09aab40cdcae5a16b723ea596620ab
 ✔ Parsed image                                                                              sha256:9ccc9244966be8bc6c3bc6f33d88a2bc062cfd21b72c055b70a33c922d09a91a
 ✔ Cataloged packages              [26573 packages]
 ✔ Scanned for vulnerabilities     [29345 vulnerability matches]
   ├── by severity: 1925 critical, 17158 high, 8845 medium, 400 low, 0 negligible (1017 unknown)
   └── by status:   24759 fixed, 4586 not-fixed, 0 ignored
```

Or, if you prefer to consume data visually:

![](severity.png)

![](installed.png)

### How it works

To minimize size, the image doesn't actually contain any packages. In fact, it only contains two files:

1. `/etc/os-release`, which tells scanners the image is a [Wolfi](https://wolfi.dev) image.
1. `/lib/apk/db/installed`, which tells scanners that the image contains every version of every package that Wolfi has available.

Wolfi aims to reduce the number of vulnerable packages by producing new fixed packages as soon as possible. But, along the way, it also produces lots and _lots_ of packages that _do_ contain vulnerabilities. This image claims to contain all of them.

Amusingly, it takes about 500ms to build and push the image, and almost two minutes to scan it.

### Why?

Aside from being fun, this image demonstrates how scanners work -- and importantly, how they _don't_ work.

At their most basic, scanners require images (1) tell them what OS they are, and (2) tell them what packages they contain. This image does both, but it does so in a way that is misleading.

For a similar (but opposite) demonstration of this, see [Malicious Compliance: Reflections on Trusting Container Scanners](https://www.youtube.com/watch?v=9weGi0csBZM).
