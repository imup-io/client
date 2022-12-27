package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

// NOTE: ClientVersion is set via build flags
var ClientVersion = "dev"

// write buffered data to the users cache directory
// used to store unsent data in the case of an unexpected shutdown
func toUserCache(data sendDataJob) {
	cache, err := os.UserCacheDir()
	if err != nil {
		log.Error(err)
	}

	targetDir := filepath.Join(cache, "imup")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		log.Error(err)
	}

	b, err := json.Marshal(data)
	if err != nil {
		log.Error(err)
	}

	h := md5.New()
	_, err = io.WriteString(h, string(b))
	if err != nil {
		log.Error(err)
	}

	if err := os.WriteFile(fmt.Sprintf("%s/%x.json", targetDir, h.Sum(nil)), b, 0666); err != nil {
		log.Error(err)
	}
}

func fromCacheDir() ([]sendDataJob, bool) {
	data := []sendDataJob{}
	cache, err := os.UserCacheDir()
	if err != nil {
		log.Error(err)
		return data, false
	}

	targetDir := filepath.Join(cache, "imup")

	// check to see if path exists before reading
	if _, err := os.Stat(targetDir); err != nil {
		log.Debug(err)
		return data, false
	}

	files, err := os.ReadDir(targetDir)
	if err != nil {
		log.Error(err)
		return data, false
	}

	for _, f := range files {
		if job, ok := fromCache(fmt.Sprintf("%s/%s", targetDir, f.Name())); ok {
			data = append(data, job)
		}
	}

	return data, len(data) > 0
}

// read imup data from users cache directory
func fromCache(name string) (sendDataJob, bool) {
	data := sendDataJob{}

	var err error
	var file *os.File
	if file, err = os.Open(name); err != nil {
		log.Error(err)
		return data, false
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return data, false
	}

	return data, true
}

func clearCache() {
	cache, err := os.UserCacheDir()
	if err != nil {
		log.Error(err)
		return
	}

	targetDir := filepath.Join(cache, "imup")
	files, err := os.ReadDir(targetDir)
	if err != nil {
		log.Error(err)
		return
	}

	for _, f := range files {
		os.Remove(fmt.Sprintf("%s/%s", targetDir, f.Name()))
	}
}
