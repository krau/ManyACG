package telegram

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"
	"github.com/krau/ManyACG/types"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

type TelegramStorage struct{}

var (
	Bot    *telego.Bot
	ChatID telego.ChatID
)

func (t *TelegramStorage) Init() {
	common.Logger.Infof("Initializing telegram storage")
	ChatID = telegoutil.ID(config.Cfg.Storage.Telegram.ChatID)
	var err error
	Bot, err = telego.NewBot(config.Cfg.Storage.Telegram.Token, telego.WithAPIServer(config.Cfg.Storage.Telegram.ApiUrl))
	if err != nil {
		common.Logger.Fatalf("failed to create telegram bot: %s", err)
		os.Exit(1)
	}
	botInfo, err := Bot.GetMe()
	if err != nil {
		common.Logger.Fatalf("failed to get bot info: %s", err)
		os.Exit(1)
	}
	common.Logger.Infof("telegram storage bot %s is ready", botInfo.Username)
}

func (t *TelegramStorage) Save(ctx context.Context, filePath string, _ string) (*types.StorageDetail, error) {
	common.Logger.Debugf("saving file %s", filePath)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	msg, err := Bot.SendDocument(telegoutil.Document(ChatID, telegoutil.File(telegoutil.NameReader(file, filepath.Base(filePath)))))
	if err != nil {
		return nil, err
	}
	fileMessage := &fileMessage{
		ChatID:     ChatID.ID,
		FileID:     msg.Document.FileID,
		MessaageID: msg.MessageID,
	}
	data, err := json.Marshal(fileMessage)
	if err != nil {
		return nil, err
	}
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
	tgFile, err := Bot.GetFile(&telego.GetFileParams{
		FileID: file.FileID,
	})
	if err != nil {
		return nil, err
	}
	return telegoutil.DownloadFile(Bot.FileDownloadURL(tgFile.FilePath))
}

func (t *TelegramStorage) Delete(ctx context.Context, detail *types.StorageDetail) error {
	// do nothing
	return nil
}
