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

func initZsvCache() bool {
	fmt.Println("initializing cache...")

	fmt.Printf("checking cache directory [%v]\n", cacheDir)
	if stat, err := os.Stat(cacheDir); err == nil && stat.IsDir() {
		fmt.Printf("cache directory exists\n")
		return true
	}

	fmt.Println("cache directory does not exist, creating...")
	if err := os.MkdirAll(cacheDir, os.ModePerm); err != nil {
		fmt.Printf("failed to create cache directory [%v]\n", err)
		return false
	}

	fmt.Printf("cache directory created [%v]\n", cacheDir)
	fmt.Println("initialized cache successfully")
	return true
}

func loadZsvCache() (map[string]int64, bool) {
	fmt.Println("loading cache")

	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		fmt.Printf("error while reading cache directory, %v\n", err)
		return nil, false
	}

	c := map[string]int64{}
	for _, e := range entries {
		name := e.Name()
		info, err := e.Info()
		if err != nil {
			fmt.Printf("failed to get directory entry info [%v], %v\n", name, err)
			return nil, false
		}
		if mode := info.Mode(); mode.IsRegular() && strings.HasSuffix(name, archive) {
			c[name] = info.Size()
		}
	}

	if len(c) == 0 {
		fmt.Printf("cache is empty\n")
		return nil, true
	}

	fmt.Printf("cache loaded successfully [%v]\n", len(c))
	return c, true
}

func cleanZsvCache(tags map[string]bool) bool {
	fmt.Printf("cleaning cache\n")

	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		fmt.Printf("failed to read cache directory, %v\n", err)
		return false
	}

	cleanList := []string{}
	for _, e := range entries {
		entryName := e.Name()
		if _, ok := tags[entryName]; !ok {
			cleanList = append(cleanList, entryName)
		}
	}

	if len(cleanList) == 0 {
		fmt.Printf("nothing to clean in cache\n")
		return true
	}

	count := 0
	fmt.Printf("clean list: %v\n", cleanList)
	for _, e := range cleanList {
		p := filepath.Join(cacheDir, e)
		if err := os.RemoveAll(p); err != nil {
			fmt.Printf("failed to remove [%v]\n", p)
		} else {
			fmt.Printf("removed %v\n", p)
			count++
		}
	}

	fmt.Printf("cleaned cache successfully [%v/%v]\n", count, len(cleanList))
	return true
}

func setupZsvCache() ([]string, error) {
	fmt.Printf("setting up zsv cache\n")

	if !initZsvCache() {
		return nil, fmt.Errorf("failed to set up cache")
	}

	c, ok := loadZsvCache()
	if !ok {
		return nil, fmt.Errorf("failed to load cache")
	}

	ctx := context.Background()
	client := github.NewClient(nil)
	opts := &github.ListOptions{Page: 1, PerPage: 3}

	fmt.Printf("checking releases\n")
	releases, _, err := client.Repositories.ListReleases(ctx, owner, repo, opts)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	versions := []string{}
	requiredDirContents := map[string]bool{}

	// tag > id
	m := map[string]int64{}

	for _, r := range releases {
		tag := r.GetTagName()
		versions = append(versions, tag)
		tarName := fmt.Sprintf("%v.%v", tag, archive)
		requiredDirContents[tag] = true
		requiredDirContents[tarName] = true
		for _, a := range r.Assets {
			if strings.HasSuffix(*a.Name, suffix) {
				id := a.GetID()
				size := a.GetSize()
				fmt.Printf("checking %v [size: %v]\n", tag, size)
				if cachedSize, ok := c[tarName]; ok && int64(size) == cachedSize {
					fmt.Printf("%v found in cache, skipped\n", tag)
				} else {
					m[tag] = id
				}
			}
		}
	}

	downloads := 0
	for tag, id := range m {
		fmt.Printf("downloading %v [id: %v]\n", tag, id)
		rc, _, err := client.Repositories.DownloadReleaseAsset(ctx, owner, repo, id, http.DefaultClient)
		if err != nil {
			fmt.Printf("failed to download %v\n", tag)
			return nil, err
		}

		filename := fmt.Sprintf("%v/%v.%v", cacheDir, tag, archive)
		file, _ := os.Create(filename)
		io.Copy(file, rc)
		file.Close()
		rc.Close()

		downloads++
	}

	if downloads > 0 {
		fmt.Printf("downloads: %v\n", downloads)
		for tag := range m {
			fmt.Printf("extracting %v\n", tag)
			filename := fmt.Sprintf("%v/%v.%v", cacheDir, tag, archive)
			file, err := os.Open(filename)
			if err != nil {
				fmt.Printf("failed to open archive [%v], %v\n", filename, err)
			}
			targetDir := filepath.Join(cacheDir, tag)
			if err := untarZsvTarGz(targetDir, file); err != nil {
				fmt.Printf("failed to untar archive [%v], %v\n", filename, err)
			}
		}
	}

	if !cleanZsvCache(requiredDirContents) {
		fmt.Printf("failed to clean cache\n")
	}

	fmt.Printf("set up zsv cache successfully\n")
	return versions, nil
}

func untarZsvTarGz(targetDir string, r io.Reader) error {
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
