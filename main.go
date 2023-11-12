package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

func main() {
	ir, err := http.Get("https://packages.wolfi.dev/os/x86_64/APKINDEX.tar.gz")
	if err != nil {
		log.Fatal(err)
	}
	if ir.StatusCode != 200 {
		log.Fatal(ir.StatusCode)
	}
	defer ir.Body.Close()

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	gr, err := gzip.NewReader(ir.Body)
	if err != nil {
		log.Fatal(err)
	}
	defer gr.Close()
	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			log.Fatal("found no APKINDEX")
		}
		if err != nil {
			log.Fatal(err)
		}
		if hdr.Name == "APKINDEX" {
			hdr.Name = "lib/apk/db/installed"

			var minimalAPKDB bytes.Buffer
			scanner := bufio.NewScanner(tr)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "P:") || strings.HasPrefix(line, "V:") || strings.HasPrefix(line, "o:") || line == "" {
					minimalAPKDB.WriteString(line + "\n")
				}
			}
			if err := scanner.Err(); err != nil {
				log.Fatal(err)
			}

			hdr.Size = int64(minimalAPKDB.Len())
			if err := tw.WriteHeader(hdr); err != nil {
				log.Fatal(err)
			}
			if _, err := tw.Write(minimalAPKDB.Bytes()); err != nil {
				log.Fatal(err)
			}
			log.Println("wrote /lib/apk/db/installed")
			break
		}
	}
	osRelease := `ID=wolfi
NAME="Wolfi"
PRETTY_NAME="Wolfi"
VERSION_ID="20230201"
HOME_URL="https://wolfi.dev"
`
	if err := tw.WriteHeader(&tar.Header{
		Name: "etc/os-release",
		Size: int64(len(osRelease)),
		Mode: 0644,
	}); err != nil {
		log.Fatal(err)
	}
	if _, err := tw.Write([]byte(osRelease)); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote /etc/os-release")

	if err := tw.Close(); err != nil {
		log.Fatal(err)
	}

	l, err := tarball.LayerFromOpener(func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(buf.Bytes())), nil
	})
	if err != nil {
		log.Fatal(err)
	}
	img, err := mutate.AppendLayers(empty.Image, l)
	if err != nil {
		log.Fatal(err)
	}
	ref := name.MustParseReference("ttl.sh/maxcve")
	if err := remote.Write(ref, img); err != nil {
		log.Fatal(err)
	}
	d, err := img.Digest()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("wrote", ref.Context().Digest(d.String()))
	fmt.Println(ref.Context().Digest(d.String()))
}
