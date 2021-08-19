// This program downloads the dist assets for the current swagger-ui version and
// places them into the embed directory.

// +build ignore

package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	const (
		embedPath   = "embed"
		indexPath   = embedPath + "/index.html"
		versionPath = "current_version.txt"
	)

	cv, err := ioutil.ReadFile("current_version.txt")
	if err != nil {
		return fmt.Errorf("couldn't read current_version.txt: %w", err)
	}
	current := string(bytes.TrimSpace(cv))

	fmt.Println("Checking latest releases...")

	latest, err := latestVersion()
	if err != nil {
		return fmt.Errorf("couldn't find latest version: %w", err)
	}

	if latest == current {
		fmt.Printf("Already up-to-date (%s)\n", latest)
		return nil
	}

	fmt.Printf("Updating from %s to %s...\n", current, latest)

	archiveURL := fmt.Sprintf("https://github.com/swagger-api/swagger-ui/archive/%s.tar.gz", latest)
	resp, err := http.Get(archiveURL)
	if err != nil {
		return fmt.Errorf("couldn't download %s: %w", archiveURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("couldn't download %s: %s", archiveURL, resp.Status)
	}

	if err := os.RemoveAll("embed"); err != nil {
		return fmt.Errorf("error removing old embed directory")
	}
	if err := os.Mkdir("embed", 0755); err != nil {
		return fmt.Errorf("error recreating embed directory")
	}

	zr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("error opening file as gzip: %w", err)
	}

	for tr := tar.NewReader(zr); ; {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar parsing error: %w", err)
		}

		// Skip everything but regular files.
		if header.Typeflag != tar.TypeReg {
			continue
		}

		filename := header.Name[strings.Index(header.Name, `/`):]
		if strings.HasPrefix(filename, `/dist`) {
			filename = strings.TrimPrefix(filename, `/dist`)
			out, err := os.Create(filepath.Join("embed", filename))
			if err != nil {
				return fmt.Errorf("couldn't extract %s: %w", filename, err)
			}
			if _, err := io.Copy(out, tr); err != nil {
				return fmt.Errorf("couldn't extract %s: %w", filename, err)
			}
		}
	}

	fmt.Printf("Rewriting %s...\n", indexPath)

	index, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("couldn't read %s: %w", indexPath, err)
	}

	index = bytes.ReplaceAll(
		index,
		[]byte(`url: "https://petstore.swagger.io/v2/swagger.json"`),
		[]byte(`url: "{{.SwaggerURL}}"`))

	if err := os.WriteFile(indexPath, index, 0644); err != nil {
		return fmt.Errorf("couldn't write index.html: %w", err)
	}

	fmt.Println("Updating version...")

	if err := os.WriteFile(versionPath, []byte(latest), 0644); err != nil {
		return fmt.Errorf("couldn't write %s: %w", versionPath, err)
	}

	fmt.Println("Done. Please run the following command to push changes.")
	fmt.Println()
	fmt.Printf("git commit -am 'Update swaggerui to %s' && git push\n", latest)
	return nil
}

// latestVersion gets the latest released version.
func latestVersion() (string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/repos/swagger-api/swagger-ui/releases", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, err := (&http.Client{Timeout: 60 * time.Second}).Do(req)
	if err != nil {
		return "", fmt.Errorf("couldn't list releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("couldn't list releases: %s", resp.Status)
	}

	var releases []struct {
		TagName    string `json:"tag_name"`
		PreRelease bool   `json:"prerelease"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	// This assumes releases are returned in descending order.
	for _, release := range releases {
		if release.PreRelease {
			// We only want stable releases.
			continue
		}

		return release.TagName, nil
	}

	return "", fmt.Errorf("no suitable releases")
}
