package cmd

import (
	"ManyACG-Bot/common"
	. "ManyACG-Bot/logger"
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"runtime"
)

const (
	GithubApiUrl string = "https://api.github.com/repos/krau/ManyACG-Bot/releases/latest"
)

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func Update() {
	resp, err := common.Client.R().Get(GithubApiUrl)
	if err != nil {
		Logger.Errorf("Error when fetching latest release: %s", err)
		return
	}
	if resp.StatusCode != 200 {
		Logger.Errorf("Error when fetching latest release: %s", resp.Status)
		return
	}
	var release githubRelease
	if err := json.Unmarshal(resp.Bytes(), &release); err != nil {
		Logger.Errorf("Error when unmarshal latest release: %s", err)
		return
	}
	if release.TagName == Version {
		Logger.Info("Already the latest version")
		return
	}
	Logger.Infof("New version %s is available", release.TagName)

	goos := runtime.GOOS
	goarch := runtime.GOARCH

	assetName := "ManyACG-" + release.TagName + "-" + goos + "-" + goarch
	if goos == "windows" {
		assetName += ".zip"
	} else {
		assetName += ".tar.gz"
	}

	var asset githubAsset
	for _, a := range release.Assets {
		if a.Name == assetName {
			asset = a
			break
		}
	}
	if asset.Name == "" {
		Logger.Errorf("No asset for %s", assetName)
		return
	}
	Logger.Infof("Downloading %s", asset.Name)
	data, err := common.DownloadFromURL(asset.BrowserDownloadURL)
	if err != nil {
		Logger.Errorf("Error when downloading asset: %s", err)
		return
	}
	if err := common.MkFile(assetName, data); err != nil {
		Logger.Errorf("Error when writing asset: %s", err)
		return
	}
	Logger.Infof("Downloaded %s", assetName)
	if goos == "windows" {
		zipFile, err := zip.OpenReader(assetName)
		if err != nil {
			Logger.Errorf("Error when opening zip file: %s", err)
			Logger.Notice("Please manually extract the file")
			return
		}
		defer zipFile.Close()
		for _, file := range zipFile.File {
			readCloser, err := file.Open()
			if err != nil {
				Logger.Errorf("Error when opening file in zip: %s", err)
				Logger.Notice("Please manually extract the file")
				return
			}
			defer readCloser.Close()
			outFile, err := os.Create(file.Name)
			if err != nil {
				Logger.Errorf("Error when creating file: %s", err)
				Logger.Notice("Please manually extract the file")
				return
			}
			defer outFile.Close()
			_, err = io.Copy(outFile, readCloser)
			if err != nil {
				Logger.Errorf("Error when copying file: %s", err)
				Logger.Notice("Please manually extract the file")
				return
			}
			Logger.Infof("Extracted %s", file.Name)
		}
		if err := common.PurgeFile(assetName); err != nil {
			Logger.Warnf("Error when purging zip file: %s", err)
		}
		Logger.Infof("Update successfully, please restart the program")
		return
	}
	tarFile, err := os.Open(assetName)
	if err != nil {
		Logger.Errorf("Error when opening tar file: %s", err)
		Logger.Notice("Please manually extract the file")
		return
	}
	defer tarFile.Close()
	gzipReader, err := gzip.NewReader(tarFile)
	if err != nil {
		Logger.Errorf("Error when opening gzip reader: %s", err)
		Logger.Notice("Please manually extract the file")
		return
	}
	defer gzipReader.Close()
	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			Logger.Errorf("Error when reading tar: %s", err)
			Logger.Notice("Please manually extract the file")
			return
		}
		outFile, err := os.Create(header.Name)
		if err != nil {
			Logger.Errorf("Error when creating file: %s", err)
			Logger.Notice("Please manually extract the file")
			return
		}
		defer outFile.Close()
		_, err = io.Copy(outFile, tarReader)
		if err != nil {
			Logger.Errorf("Error when copying file: %s", err)
			Logger.Notice("Please manually extract the file")
			return
		}
		Logger.Infof("Extracted %s", header.Name)
	}
	if err := common.PurgeFile(assetName); err != nil {
		Logger.Warnf("Error when purging tar file: %s", err)
	}
	Logger.Infof("Update successfully, please restart the program")
}
