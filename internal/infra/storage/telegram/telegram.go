package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"
	"github.com/krau/ManyACG/types"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoapi"
	"github.com/mymmrac/telego/telegoutil"
)

type TelegramStorage struct{}

var (
	Bot    *telego.Bot
	ChatID telego.ChatID
)

func (t *TelegramStorage) Init(ctx context.Context) {
	common.Logger.Infof("Initializing telegram storage")
	ChatID = telegoutil.ID(config.Cfg.Storage.Telegram.ChatID)
	var err error
	Bot, err = telego.NewBot(config.Cfg.Storage.Telegram.Token, telego.WithAPIServer(config.Cfg.Storage.Telegram.ApiUrl),
		telego.WithAPICaller(&telegoapi.RetryCaller{
			Caller:       telegoapi.DefaultFastHTTPCaller,
			MaxAttempts:  config.Cfg.Storage.Telegram.Retry.MaxAttempts,
			ExponentBase: config.Cfg.Storage.Telegram.Retry.ExponentBase,
			StartDelay:   time.Duration(config.Cfg.Storage.Telegram.Retry.StartDelay) * time.Second,
			MaxDelay:     time.Duration(config.Cfg.Storage.Telegram.Retry.MaxDelay) * time.Second,
			RateLimit:    telegoapi.RetryRateLimitWaitOrAbort,
		}))
	if err != nil {
		common.Logger.Panicf("failed to create telegram bot: %s", err)
	}
	botInfo, err := Bot.GetMe(ctx)
	if err != nil {
		common.Logger.Panicf("failed to get bot info: %s", err)
	}
	common.Logger.Infof("telegram storage bot %s is ready", botInfo.Username)
}

func (t *TelegramStorage) Save(ctx context.Context, filePath string, _ string) (*types.StorageDetail, error) {
	common.Logger.Debugf("saving file %s", filePath)
	var msg *telego.Message
	var err error
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		common.Logger.Errorf("failed to read file: %s", err)
		return nil, ErrReadFile
	}
	for i := range config.Cfg.Storage.Telegram.Retry.MaxAttempts {
		msg, err = Bot.SendDocument(ctx, telegoutil.Document(ChatID, telegoutil.File(telegoutil.NameReader(bytes.NewReader(fileBytes), filepath.Base(filePath)))))
		if err != nil {
			var apiErr *telegoapi.Error
			if errors.As(err, &apiErr) && apiErr.ErrorCode == 429 && apiErr.Parameters != nil {
				retryAfter := apiErr.Parameters.RetryAfter + (i * int(config.Cfg.Storage.Telegram.Retry.StartDelay))
				common.Logger.Warnf("Rate limited, retry after %d seconds", retryAfter)
				time.Sleep(time.Duration(retryAfter) * time.Second)
				continue
			}
			common.Logger.Errorf("failed to send document: %s", err)
			return nil, fmt.Errorf("failed to send document: %w", err)
		}
		break
	}
	if err != nil {
		common.Logger.Errorf("failed to send document: %s", err)
		return nil, fmt.Errorf("failed to send document: %w", err)
	}
	fileMessage := &fileMessage{
		ChatID:     ChatID.ID,
		FileID:     msg.Document.FileID,
		MessaageID: msg.MessageID,
	}
	data, err := json.Marshal(fileMessage)
	if err != nil {
		common.Logger.Errorf("failed to marshal file message: %s", err)
		return nil, ErrFailedMarshalFileMessage
	}
	cachePath := filepath.Join(config.Cfg.Storage.CacheDir, common.MD5Hash(fileMessage.FileID))
	go common.MkCache(cachePath, fileBytes, time.Duration(config.Cfg.Storage.CacheTTL)*time.Second)
	return &types.StorageDetail{
		Type: types.StorageTypeTelegram,
		Path: string(data),
	}, nil
}

func (t *TelegramStorage) GetFile(ctx context.Context, detail *types.StorageDetail) ([]byte, error) {
	var file fileMessage
	if err := json.Unmarshal([]byte(detail.Path), &file); err != nil {
		return nil, err
	}
	common.Logger.Debugf("getting file %s", file.String())
	cachePath := filepath.Join(config.Cfg.Storage.CacheDir, common.MD5Hash(file.FileID))
	if data, err := os.ReadFile(cachePath); err == nil {
		return data, nil
	}
	tgFile, err := Bot.GetFile(ctx, &telego.GetFileParams{
		FileID: file.FileID,
	})
	if err != nil {
		return nil, err
	}
	data, err := telegoutil.DownloadFile(Bot.FileDownloadURL(tgFile.FilePath))
	if err != nil {
		return nil, err
	}
	go common.MkCache(cachePath, data, time.Duration(config.Cfg.Storage.CacheTTL)*time.Second)
	return data, nil
}

func (t *TelegramStorage) Delete(ctx context.Context, detail *types.StorageDetail) error {
	// do nothing
	return nil
}
