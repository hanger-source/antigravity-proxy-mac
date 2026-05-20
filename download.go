package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const singboxVersion = "1.11.15"

func downloadSingBox(destDir string) (string, error) {
	binPath := filepath.Join(destDir, "sing-box")

	arch := runtime.GOARCH
	if arch == "arm64" {
		arch = "arm64"
	} else {
		arch = "amd64"
	}

	url := fmt.Sprintf("https://github.com/SagerNet/sing-box/releases/download/v%s/sing-box-%s-darwin-%s.tar.gz",
		singboxVersion, singboxVersion, arch)

	logInfo("downloading sing-box v%s from %s", singboxVersion, url)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("download failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("gzip error: %v", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("tar error: %v", err)
		}
		if strings.HasSuffix(hdr.Name, "/sing-box") && hdr.Typeflag == tar.TypeReg {
			f, err := os.OpenFile(binPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
			if err != nil {
				return "", err
			}
			io.Copy(f, tr)
			f.Close()
			logInfo("sing-box v%s installed to %s", singboxVersion, binPath)
			return binPath, nil
		}
	}

	return "", fmt.Errorf("sing-box binary not found in archive")
}
