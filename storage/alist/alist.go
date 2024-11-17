package alist

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"

	"github.com/krau/ManyACG/types"

	"github.com/imroc/req/v3"
)

type Alist struct{}

var (
	basePath string
	baseUrl  string
)

var (
	reqClient *req.Client
	loginReq  *loginRequset
)

func (a *Alist) Init() {
	alistConfig := config.Cfg.Storage.Alist
	basePath = strings.TrimSuffix(alistConfig.Path, "/")
	baseUrl = strings.TrimSuffix(alistConfig.URL, "/")
	reqClient = req.C().
		SetCommonRetryCount(2).
		SetTLSHandshakeTimeout(time.Second * 10).
		SetBaseURL(baseUrl)
	loginReq = &loginRequset{
		Username: alistConfig.Username,
		Password: alistConfig.Password,
	}
	token, err := getJwtToken()
	if err != nil {
		common.Logger.Errorf("Failed to login to Alist: %v", err)
		os.Exit(1)
	}
	common.Logger.Debugf("Login to Alist successfully")
	reqClient.SetCommonHeader("Authorization", token)
	go refreshJwtToken(reqClient)
}

func (a *Alist) Save(ctx context.Context, filePath string, storagePath string) (*types.StorageDetail, error) {
	common.Logger.Debugf("saving file %s", filePath)
	storagePath = path.Join(basePath, storagePath)
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		common.Logger.Errorf("failed to read file: %s", err)
		return nil, err
	}
	resp, err := reqClient.R().SetContext(ctx).SetFileBytes("file", filepath.Base(storagePath), fileBytes).
		SetHeaders(map[string]string{
			"File-Path": url.PathEscape(storagePath),
			"As-Task":   "true",
		}).Put("/api/fs/form")
	if err != nil {
		common.Logger.Errorf("failed to save file: %s", err)
		return nil, err
	}
	var fsFormResp FsFormResponse
	if err := json.Unmarshal(resp.Bytes(), &fsFormResp); err != nil {
		common.Logger.Errorf("failed to unmarshal response: %s", err)
		return nil, err
	}
	if fsFormResp.Code != http.StatusOK {
		common.Logger.Errorf("failed to save file: %s", fsFormResp.Message)
		return nil, fmt.Errorf("failed to save file: %s", fsFormResp.Message)
	}
	cachePath := strings.TrimSuffix(config.Cfg.Storage.CacheDir, "/") + "/" + filepath.Base(storagePath)
	go common.MkCache(cachePath, fileBytes, time.Duration(config.Cfg.Storage.CacheTTL)*time.Second)
	return &types.StorageDetail{
		Type: types.StorageTypeAlist,
		Path: storagePath,
	}, nil
}

func (a *Alist) GetFile(ctx context.Context, detail *types.StorageDetail) ([]byte, error) {
	common.Logger.Debugf("Getting file %s", detail.Path)
	cachePath := path.Join(config.Cfg.Storage.CacheDir, filepath.Base(detail.Path))
	data, err := os.ReadFile(cachePath)
	if err == nil {
		return data, nil
	}
	resp, err := reqClient.R().SetContext(ctx).SetBodyJsonMarshal(map[string]string{
		"path":     detail.Path,
		"password": config.Cfg.Storage.Alist.PathPassword,
	}).Post("/api/fs/get")
	if err != nil {
		common.Logger.Errorf("failed to get file: %s", err)
		return nil, err
	}
	var fsGetResp FsGetResponse
	if err := json.Unmarshal(resp.Bytes(), &fsGetResp); err != nil {
		common.Logger.Errorf("failed to unmarshal response: %s", err)
		return nil, err
	}
	if fsGetResp.Code != http.StatusOK {
		common.Logger.Errorf("failed to get file: %s", fsGetResp.Message)
		return nil, fmt.Errorf("failed to get file: %s", fsGetResp.Message)
	}
	_, err = reqClient.R().SetContext(ctx).SetOutputFile(cachePath).Get(fsGetResp.Data.RawUrl)
	if err != nil {
		common.Logger.Errorf("failed to save file: %s", err)
		return nil, err
	}
	go common.PurgeFileAfter(cachePath, time.Duration(config.Cfg.Storage.CacheTTL)*time.Second)
	return os.ReadFile(cachePath)
}

func (a *Alist) Delete(ctx context.Context, detail *types.StorageDetail) error {
	common.Logger.Debugf("Deleting file %s", detail.Path)
	resp, err := reqClient.R().SetContext(ctx).SetBodyJsonMarshal(map[string]any{
		"names": []string{filepath.Base(detail.Path)},
		"dir":   filepath.Dir(detail.Path),
	}).Post("/api/fs/remove")
	if err != nil {
		common.Logger.Errorf("failed to delete file: %s", err)
		return err
	}
	var fsRemoveResp FsRemoveResponse
	if err := json.Unmarshal(resp.Bytes(), &fsRemoveResp); err != nil {
		common.Logger.Errorf("failed to unmarshal response: %s", err)
		return err
	}
	if fsRemoveResp.Code != http.StatusOK {
		common.Logger.Errorf("failed to delete file: %s", fsRemoveResp.Message)
		return fmt.Errorf("failed to delete file: %s", fsRemoveResp.Message)
	}
	return nil
}
