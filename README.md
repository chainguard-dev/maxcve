# _MAXIMUM CVEs_

[![Release](https://github.com/imjasonh/maxcve/actions/workflows/release.yaml/badge.svg)](https://github.com/imjasonh/maxcve/actions/workflows/release.yaml)

This repo generates a container image that maximizes the number of CVEs in the image, while minimizing the size of the image.

The result is a 183 KB image that it has _more than 35,000 known vulnerabilities_. That's roughly one CVE for every 5 bytes of image data!

```
grype ghcr.io/imjasonh/maxcve/maxcve 1> /dev/null
...
 ✔ Vulnerability DB                [updated]  
 ✔ Cataloged contents
   └── ✔ Packages                        [45,598 packages]  
 ✔ Scanned for vulnerabilities     [249238 vulnerability matches]  
   ├── by severity: 5080 critical, 47441 high, 34209 medium, 1386 low, 0 negligible (161122 unknown)
   └── by status:   241041 fixed, 8197 not-fixed, 0 ignored 
```

(As of March 14, 2024)

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
