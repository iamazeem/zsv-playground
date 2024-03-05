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
	"strings"

	"github.com/google/go-github/v60/github"
)

func initZsvCache() bool {
	log.Println("initializing cache")

	log.Printf("checking cache directory [%v]\n", cacheDir)
	if stat, err := os.Stat(cacheDir); err == nil && stat.IsDir() {
		log.Printf("cache directory exists\n")
		return true
	}

	log.Println("cache directory does not exist, creating")
	if err := os.MkdirAll(cacheDir, os.ModePerm); err != nil {
		log.Printf("failed to create cache directory [%v]\n", err)
		return false
	}

	log.Printf("cache directory created [%v]\n", cacheDir)
	log.Println("initialized cache successfully")
	return true
}

func loadZsvCache() (map[string]int64, bool) {
	log.Println("loading cache")

	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		log.Printf("error while reading cache directory, %v\n", err)
		return nil, false
	}

	cache := map[string]int64{}
	for _, e := range entries {
		name := e.Name()
		info, err := e.Info()
		if err != nil {
			log.Printf("failed to get directory entry info [%v], %v\n", name, err)
			return nil, false
		}
		if mode := info.Mode(); mode.IsRegular() && strings.HasSuffix(name, archive) {
			cache[name] = info.Size()
		}
	}

	if len(cache) == 0 {
		log.Printf("cache is empty\n")
		return nil, true
	}

	log.Printf("cache loaded successfully [%v]\n", len(cache))
	return cache, true
}

func cleanZsvCache(tags map[string]bool) bool {
	log.Printf("cleaning cache\n")

	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		log.Printf("failed to read cache directory, %v\n", err)
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
		log.Printf("nothing to clean in cache\n")
		return true
	}

	count := 0
	log.Printf("clean list: %v\n", cleanList)
	for _, e := range cleanList {
		p := filepath.Join(cacheDir, e)
		if err := os.RemoveAll(p); err != nil {
			log.Printf("failed to remove [%v]\n", p)
		} else {
			log.Printf("removed %v\n", p)
			count++
		}
	}

	log.Printf("cleaned cache successfully [%v/%v]\n", count, len(cleanList))
	return true
}

func setupZsvCache() ([]string, error) {
	log.Printf("setting up zsv cache\n")

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

	log.Printf("checking latest zsv releases [%v]\n", opts.Page*opts.PerPage)
	releases, _, err := client.Repositories.ListReleases(ctx, owner, repo, opts)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	versions := []string{}
	requiredCacheContents := map[string]bool{}

	// tag > id
	m := map[string]int64{}

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
				log.Printf("checking %v [size: %v]\n", tag, size)
				if cachedSize, ok := c[tarName]; ok && int64(size) == cachedSize {
					log.Printf("%v found in cache, skipped\n", tag)
				} else {
					m[tag] = id
				}
			}
		}
	}

	downloads := 0
	for tag, id := range m {
		log.Printf("downloading %v [id: %v]\n", tag, id)
		rc, _, err := client.Repositories.DownloadReleaseAsset(ctx, owner, repo, id, http.DefaultClient)
		if err != nil {
			log.Printf("failed to download %v\n", tag)
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
		log.Printf("downloads: %v\n", downloads)
		for tag := range m {
			log.Printf("extracting %v\n", tag)
			filename := fmt.Sprintf("%v/%v.%v", cacheDir, tag, archive)
			file, err := os.Open(filename)
			if err != nil {
				log.Printf("failed to open archive [%v], %v\n", filename, err)
			}
			targetDir := filepath.Join(cacheDir, tag)
			if err := untarZsvTarGz(targetDir, file); err != nil {
				log.Printf("failed to untar archive [%v], %v\n", filename, err)
			}
		}
	}

	if !cleanZsvCache(requiredCacheContents) {
		log.Printf("failed to clean cache\n")
	}

	log.Printf("set up zsv cache successfully\n")
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

func generateZsvExePath(version string) string {
	return fmt.Sprintf("%v/%v/%v/bin/zsv", cacheDir, version, triplet)
}

func getZsvExePaths(versions []string) []string {
	bins := []string{}
	for _, v := range versions {
		zsv := generateZsvExePath(v)
		bins = append(bins, zsv)
	}
	return bins
}
