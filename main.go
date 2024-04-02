package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

func main() {
	dst := "ttl.sh/maxcve"
	if len(os.Args) > 1 {
		dst = os.Args[1]
	}

	ir, err := http.Get("https://packages.wolfi.dev/os/x86_64/APKINDEX.json")
	if err != nil {
		log.Fatal(err)
	}
	defer ir.Body.Close()
	body, err := io.ReadAll(ir.Body)
	if err != nil {
		log.Fatal(err)
	}
	if ir.StatusCode != 200 {
		log.Fatal(ir.StatusCode)
	}

	var index struct {
		Packages []struct {
			Name    string
			Version string
			Origin  string
		}
	}
	if err := json.Unmarshal(body, &index); err != nil {
		log.Fatal(err)
	}

	sort.Slice(index.Packages, func(idx, jdx int) bool {
		if index.Packages[idx].Name == index.Packages[jdx].Name {
			return index.Packages[idx].Version < index.Packages[jdx].Version
		}
		return index.Packages[idx].Name < index.Packages[jdx].Name
	})

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	var minimalAPKDB bytes.Buffer
	for _, pkg := range index.Packages {
		minimalAPKDB.WriteString("P:" + pkg.Name + "\n")
		minimalAPKDB.WriteString("V:" + pkg.Version + "\n")
		if pkg.Origin != "" {
			minimalAPKDB.WriteString("o:" + pkg.Origin + "\n")
		}
		minimalAPKDB.WriteString("\n")
	}

	if err := tw.WriteHeader(&tar.Header{
		Name: "lib/apk/db/installed",
		Size: int64(minimalAPKDB.Len()),
		Mode: 0644,
	}); err != nil {
		log.Fatal(err)
	}
	if _, err := tw.Write(minimalAPKDB.Bytes()); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote /lib/apk/db/installed")

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
	}, tarball.WithCompressionLevel(gzip.BestCompression))
	if err != nil {
		log.Fatal(err)
	}
	img, err := mutate.AppendLayers(empty.Image, l)
	if err != nil {
		log.Fatal(err)
	}
	ref, err := name.ParseReference(dst)
	if err != nil {
		log.Fatal(err)
	}
	if err := remote.Write(ref, img,
		remote.WithAuthFromKeychain(authn.DefaultKeychain)); err != nil {
		log.Fatal(err)
	}
	d, err := img.Digest()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("wrote", ref.Context().Digest(d.String()))
	fmt.Println(ref.Context().Digest(d.String()))
}
