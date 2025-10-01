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

	config "github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/storage"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/osutil"
	"github.com/krau/ManyACG/pkg/strutil"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoapi"
	"github.com/mymmrac/telego/telegoutil"
)

type TelegramStorage struct {
	cfg    config.StorageTelegramConfig
	bot    *telego.Bot
	chatID telego.ChatID
}

func init() {
	storageCfg := config.Get().Storage.Telegram
	storageType := shared.StorageTypeTelegram
	storage.Register(storageType, func() storage.Storage {
		return &TelegramStorage{
			cfg: storageCfg,
		}
	})
}

func (t *TelegramStorage) Init(ctx context.Context) error {
	t.chatID = telegoutil.ID(t.cfg.ChatID)
	var err error
	t.bot, err = telego.NewBot(t.cfg.Token, telego.WithAPIServer(t.cfg.ApiUrl),
		telego.WithAPICaller(&telegoapi.RetryCaller{
			Caller:       telegoapi.DefaultFastHTTPCaller,
			MaxAttempts:  t.cfg.Retry.MaxAttempts,
			ExponentBase: t.cfg.Retry.ExponentBase,
			StartDelay:   time.Duration(t.cfg.Retry.StartDelay) * time.Second,
			MaxDelay:     time.Duration(t.cfg.Retry.MaxDelay) * time.Second,
			RateLimit:    telegoapi.RetryRateLimitWaitOrAbort,
		}))
	if err != nil {
		return fmt.Errorf("failed to create telegram bot: %w", err)
	}
	_, err = t.bot.GetMe(ctx)
	if err != nil {
		return fmt.Errorf("failed to get bot info: %w", err)
	}
	return nil
}

func (t *TelegramStorage) Save(ctx context.Context, filePath string, _ string) (*shared.StorageDetail, error) {
	var msg *telego.Message
	var err error
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, ErrReadFile
	}
	for i := range t.cfg.Retry.MaxAttempts {
		msg, err = t.bot.SendDocument(ctx, telegoutil.Document(t.chatID, telegoutil.File(telegoutil.NameReader(bytes.NewReader(fileBytes), filepath.Base(filePath)))))
		if err != nil {
			var apiErr *telegoapi.Error
			if errors.As(err, &apiErr) && apiErr.ErrorCode == 429 && apiErr.Parameters != nil {
				retryAfter := apiErr.Parameters.RetryAfter + (i * int(t.cfg.Retry.StartDelay))
				time.Sleep(time.Duration(retryAfter) * time.Second)
				continue
			}
			return nil, fmt.Errorf("failed to send document: %w", err)
		}
		break
	}
	if err != nil {
		return nil, fmt.Errorf("failed to send document: %w", err)
	}
	fileMessage := &fileMessage{
		ChatID:     t.chatID.ID,
		FileID:     msg.Document.FileID,
		MessaageID: msg.MessageID,
	}
	data, err := json.Marshal(fileMessage)
	if err != nil {
		return nil, ErrFailedMarshalFileMessage
	}
	cachePath := filepath.Join(config.Get().Storage.CacheDir, strutil.MD5Hash(fileMessage.FileID))
	go osutil.MkCache(cachePath, fileBytes, time.Duration(config.Get().Storage.CacheTTL)*time.Second)
	return &shared.StorageDetail{
		Type: shared.StorageTypeTelegram,
		Path: string(data),
	}, nil
}

func (t *TelegramStorage) GetFile(ctx context.Context, detail *shared.StorageDetail) ([]byte, error) {
	var file fileMessage
	if err := json.Unmarshal([]byte(detail.Path), &file); err != nil {
		return nil, err
	}
	cachePath := filepath.Join(config.Get().Storage.CacheDir, strutil.MD5Hash(file.FileID))
	if data, err := os.ReadFile(cachePath); err == nil {
		return data, nil
	}
	tgFile, err := t.bot.GetFile(ctx, &telego.GetFileParams{
		FileID: file.FileID,
	})
	if err != nil {
		return nil, err
	}
	data, err := telegoutil.DownloadFile(t.bot.FileDownloadURL(tgFile.FilePath))
	if err != nil {
		return nil, err
	}
	go osutil.MkCache(cachePath, data, time.Duration(config.Get().Storage.CacheTTL)*time.Second)
	return data, nil
}

func (t *TelegramStorage) Delete(ctx context.Context, detail *shared.StorageDetail) error {
	// do nothing
	return nil
}
