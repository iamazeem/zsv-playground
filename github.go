package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v60/github"
)

func setupCache() bool {
	fmt.Println("setting up cache")

	fmt.Printf("checking existing cache directory [%v]\n", cacheDir)
	if stat, err := os.Stat(cacheDir); err == nil && stat.IsDir() {
		fmt.Println("cache directory exists")
		return true
	}

	fmt.Println("cache directory not found, creating...")
	if err := os.MkdirAll(cacheDir, os.ModePerm); err != nil {
		fmt.Printf("failed to create cache directory [%v]\n", err)
		return false
	}

	fmt.Println("cache directory set up successfully")
	return true
}

func loadCache() (map[string]int64, bool) {
	fmt.Println("loading cache")

	c := map[string]int64{}

	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		fmt.Printf("error while reading cache directory, %v\n", err)
		return nil, false
	}

	for _, e := range entries {
		name := e.Name()
		info, err := e.Info()
		if err != nil {
			fmt.Printf("failed to get directory entry info [%v], %v\n", name, err)
			return nil, false
		}
		c[name] = info.Size()
	}

	fmt.Println("cache loaded successfully")
	return c, true
}

func downloadLatestRelease() error {
	if !setupCache() {
		return fmt.Errorf("failed to set up cache")
	}

	c, ok := loadCache()
	if !ok {
		return fmt.Errorf("failed to load cache")
	}

	ctx := context.Background()
	client := github.NewClient(nil)
	opts := &github.ListOptions{Page: 1, PerPage: 3}

	fmt.Println("listing releases...")
	releases, _, err := client.Repositories.ListReleases(ctx, owner, repo, opts)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// TODO: Purge previous cached releases in case of new ones
	// fmt.Println(len(releases))

	// tag > id
	m := map[string]int64{}

	for _, r := range releases {
		tag := r.GetTagName()
		for _, a := range r.Assets {
			if strings.HasSuffix(*a.Name, suffix) {
				id := a.GetID()
				size := a.GetSize()
				fmt.Printf("checking %v [size: %v]\n", tag, size)
				if cachedSize, ok := c[tag+".tar.gz"]; ok && int64(size) == cachedSize {
					fmt.Printf("%v found in cache, skipped\n", tag)
				} else {
					m[tag] = id
				}
			}
		}
	}

	// fmt.Println(m)
	downloads := 0
	for tag, id := range m {
		fmt.Printf("downloading %v [id: %v]\n", tag, id)
		rc, _, err := client.Repositories.DownloadReleaseAsset(ctx, owner, repo, id, http.DefaultClient)
		if err != nil {
			fmt.Printf("failed to download %v\n", tag)
			return err
		}

		filename := fmt.Sprintf("zsv/%v.tar.gz", tag)
		file, _ := os.Create(filename)
		io.Copy(file, rc)

		downloads++

		file.Close()
		rc.Close()
	}

	fmt.Printf("extracting new archives [%v]\n", downloads)
	for tag := range m {
		fmt.Printf("extracting %v\n", tag)
		filename := fmt.Sprintf("zsv/%v.tar.gz", tag)
		file, err := os.Open(filename)
		if err != nil {
			fmt.Printf("failed to open archive [%v], %v\n", filename, err)
		}
		targetDir := filepath.Join(cacheDir, tag)
		if err := untar(targetDir, file); err != nil {
			fmt.Printf("failed to untar archive [%v], %v\n", filename, err)
		}
	}

	fmt.Printf("cache populated successfully [downloads: %v]\n", downloads)
	return nil
}

func untar(targetDir string, r io.Reader) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		// extract binary only i.e. .../bin/zsv
		if !strings.HasSuffix(header.Name, "bin/") && !strings.HasSuffix(header.Name, "/bin/zsv") {
			continue
		}

		target := filepath.Join(targetDir, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}
			f.Close()
		}
	}
}
