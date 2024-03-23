package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/go-github/v60/github"
)

const (
	archive  = "tar.gz"
	cacheDir = "zsv"
)

func getTriplet() string {
	switch runtime.GOOS {
	case "linux":
		return "amd64-linux-gcc"
	case "windows":
		return "amd64-windows-mingw"
	case "darwin":
		return "amd64-macosx-gcc"
	case "freebsd":
		return "amd64-freebsd-gcc"
	default:
		log.Printf("%v not supported", runtime.GOOS)
		return ""
	}
}

func initCache() bool {
	log.Println("initializing cache")

	log.Printf("checking cache directory [%v]", cacheDir)
	if stat, err := os.Stat(cacheDir); err == nil && stat.IsDir() {
		log.Printf("cache directory exists")
		return true
	}

	log.Println("cache directory does not exist, creating")
	if err := os.MkdirAll(cacheDir, os.ModePerm); err != nil {
		log.Printf("failed to create cache directory [%v]", err)
		return false
	}

	log.Printf("cache directory created [%v]", cacheDir)
	log.Println("initialized cache successfully")
	return true
}

func loadCache() (map[string]int64, bool) {
	log.Println("loading cache")

	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		log.Printf("error while reading cache directory, %v", err)
		return nil, false
	}

	cache := map[string]int64{}
	for _, e := range entries {
		name := e.Name()
		info, err := e.Info()
		if err != nil {
			log.Printf("failed to get directory entry info [%v], %v", name, err)
			return nil, false
		}
		if mode := info.Mode(); mode.IsRegular() && strings.HasSuffix(name, archive) {
			cache[name] = info.Size()
		}
	}

	if len(cache) == 0 {
		log.Printf("cache is empty")
		return nil, true
	}

	log.Printf("cache loaded successfully [%v]", len(cache))
	return cache, true
}

func cleanCache(tags map[string]bool) bool {
	log.Printf("cleaning cache")

	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		log.Printf("failed to read cache directory, %v", err)
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
		log.Printf("nothing to clean in cache")
		return true
	}

	count := 0
	log.Printf("clean list: %v", cleanList)
	for _, e := range cleanList {
		p := filepath.Join(cacheDir, e)
		if err := os.RemoveAll(p); err != nil {
			log.Printf("failed to remove [%v]", p)
		} else {
			log.Printf("removed %v", p)
			count++
		}
	}

	log.Printf("cleaned cache successfully [%v/%v]", count, len(cleanList))
	return true
}

func setupCache() ([]string, error) {
	log.Printf("setting up zsv cache")

	if !initCache() {
		return nil, fmt.Errorf("failed to set up cache")
	}

	c, ok := loadCache()
	if !ok {
		return nil, fmt.Errorf("failed to load cache")
	}

	owner := "liquidaty"
	repo := "zsv"

	ctx := context.Background()
	client := github.NewClient(nil)
	opts := &github.ListOptions{Page: 1, PerPage: 3}

	log.Printf("checking latest zsv releases [%v]", opts.Page*opts.PerPage)
	releases, _, err := client.Repositories.ListReleases(ctx, owner, repo, opts)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	versions := []string{}
	requiredCacheContents := map[string]bool{}

	// tag > id
	m := map[string]int64{}

	suffix := getTriplet() + "." + archive
	for _, r := range releases {
		tag := r.GetTagName()
		versions = append(versions, tag)
		tarName := fmt.Sprintf("%v.%v", tag, archive)
		requiredCacheContents[tag] = true
		requiredCacheContents[tarName] = true
		for _, a := range r.Assets {
			if strings.HasSuffix(*a.Name, suffix) {
				id := a.GetID()
				size := a.GetSize()
				log.Printf("checking %v [size: %v]", tag, size)
				if cachedSize, ok := c[tarName]; ok && int64(size) == cachedSize {
					log.Printf("%v found in cache, skipped", tag)
				} else {
					m[tag] = id
				}
			}
		}
	}

	downloads := 0
	for tag, id := range m {
		log.Printf("downloading %v [id: %v]", tag, id)
		rc, _, err := client.Repositories.DownloadReleaseAsset(ctx, owner, repo, id, http.DefaultClient)
		if err != nil {
			log.Printf("failed to download %v", tag)
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
		log.Printf("downloads: %v", downloads)
		for tag := range m {
			log.Printf("extracting %v", tag)
			filename := fmt.Sprintf("%v/%v.%v", cacheDir, tag, archive)
			file, err := os.Open(filename)
			if err != nil {
				log.Printf("failed to open archive [%v], %v", filename, err)
			}
			targetDir := filepath.Join(cacheDir, tag)
			if err := untarZsvTarGz(targetDir, file); err != nil {
				log.Printf("failed to untar archive [%v], %v", filename, err)
			}
		}
	}

	if !cleanCache(requiredCacheContents) {
		log.Printf("failed to clean cache")
	}

	log.Printf("cached zsv versions: %v", versions)

	zsvExePaths := getExePaths(versions)
	log.Printf("cached zsv binaries: %v", zsvExePaths)

	log.Printf("set up zsv cache successfully")
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

func getExePath(version string) string {
	return fmt.Sprintf("%v/%v/%v/bin/zsv", cacheDir, version, getTriplet())
}

func getExePaths(versions []string) []string {
	bins := []string{}
	for _, v := range versions {
		zsv := getExePath(v)
		bins = append(bins, zsv)
	}
	return bins
}
