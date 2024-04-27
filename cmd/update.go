package cmd

import (
	"ManyACG-Bot/common"
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"runtime"

	"github.com/gookit/slog"
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
	slog.DefaultChannelName = "ManyACG-Bot"
	resp, err := common.Client.R().Get(GithubApiUrl)
	if err != nil {
		slog.Errorf("Error when fetching latest release: %s", err)
		return
	}
	if resp.StatusCode != 200 {
		slog.Errorf("Error when fetching latest release: %s", resp.Status)
		return
	}
	var release githubRelease
	if err := json.Unmarshal(resp.Bytes(), &release); err != nil {
		slog.Errorf("Error when unmarshal latest release: %s", err)
		return
	}
	if release.TagName == Version {
		slog.Info("Already the latest version")
		return
	}
	slog.Infof("New version %s is available", release.TagName)

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
		slog.Errorf("No asset for %s", assetName)
		return
	}
	slog.Infof("Downloading %s", asset.Name)
	data, err := common.DownloadFromURL(asset.BrowserDownloadURL)
	if err != nil {
		slog.Errorf("Error when downloading asset: %s", err)
		return
	}
	if err := common.MkFile(assetName, data); err != nil {
		slog.Errorf("Error when writing asset: %s", err)
		return
	}
	slog.Infof("Downloaded %s", assetName)
	if goos == "windows" {
		zipFile, err := zip.OpenReader(assetName)
		if err != nil {
			slog.Errorf("Error when opening zip file: %s", err)
			slog.Notice("Please manually extract the file")
			return
		}
		defer zipFile.Close()
		for _, file := range zipFile.File {
			readCloser, err := file.Open()
			if err != nil {
				slog.Errorf("Error when opening file in zip: %s", err)
				slog.Notice("Please manually extract the file")
				return
			}
			defer readCloser.Close()
			outFile, err := os.Create(file.Name)
			if err != nil {
				slog.Errorf("Error when creating file: %s", err)
				slog.Notice("Please manually extract the file")
				return
			}
			defer outFile.Close()
			_, err = io.Copy(outFile, readCloser)
			if err != nil {
				slog.Errorf("Error when copying file: %s", err)
				slog.Notice("Please manually extract the file")
				return
			}
			slog.Infof("Extracted %s", file.Name)
		}
		if err := common.PurgeFile(assetName); err != nil {
			slog.Warnf("Error when purging zip file: %s", err)
		}
		slog.Infof("Update successfully, please restart the program")
		return
	}
	tarFile, err := os.Open(assetName)
	if err != nil {
		slog.Errorf("Error when opening tar file: %s", err)
		slog.Notice("Please manually extract the file")
		return
	}
	defer tarFile.Close()
	gzipReader, err := gzip.NewReader(tarFile)
	if err != nil {
		slog.Errorf("Error when opening gzip reader: %s", err)
		slog.Notice("Please manually extract the file")
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
			slog.Errorf("Error when reading tar: %s", err)
			slog.Notice("Please manually extract the file")
			return
		}
		outFile, err := os.Create(header.Name)
		if err != nil {
			slog.Errorf("Error when creating file: %s", err)
			slog.Notice("Please manually extract the file")
			return
		}
		defer outFile.Close()
		_, err = io.Copy(outFile, tarReader)
		if err != nil {
			slog.Errorf("Error when copying file: %s", err)
			slog.Notice("Please manually extract the file")
			return
		}
		slog.Infof("Extracted %s", header.Name)
	}
	if err := common.PurgeFile(assetName); err != nil {
		slog.Warnf("Error when purging tar file: %s", err)
	}
	slog.Infof("Update successfully, please restart the program")
}
