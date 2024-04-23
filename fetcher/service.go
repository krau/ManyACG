package fetcher

import (
	"ManyACG-Bot/errors"
	"ManyACG-Bot/service"
	"ManyACG-Bot/storage"
	"ManyACG-Bot/telegram"
	"ManyACG-Bot/types"
	"context"
	es "errors"
	"fmt"

	. "ManyACG-Bot/logger"

	"github.com/mymmrac/telego"
	"go.mongodb.org/mongo-driver/mongo"
)

func PostAndCreateArtwork(ctx context.Context, artwork *types.Artwork, bot *telego.Bot, storage storage.Storage) error {
	artworkInDB, err := service.GetArtworkByURL(ctx, artwork.SourceURL)
	if err == nil && artworkInDB != nil {
		Logger.Debugf("Artwork %s already exists", artwork.Title)
		return errors.ErrArtworkAlreadyExist
	}
	if err != nil && !es.Is(err, mongo.ErrNoDocuments) {
		return err
	}
	messages, err := telegram.PostArtwork(telegram.Bot, artwork)
	if err != nil {
		return fmt.Errorf("posting artwork [%s](%s): %w", artwork.Title, artwork.SourceURL, err)
	}
	Logger.Infof("Posted artwork %s", artwork.Title)

	for i, picture := range artwork.Pictures {
		if picture.TelegramInfo == nil {
			picture.TelegramInfo = &types.TelegramInfo{}
		}
		picture.TelegramInfo.MessageID = messages[i].MessageID
		picture.TelegramInfo.MediaGroupID = messages[i].MediaGroupID
		if messages[i].Photo != nil {
			picture.TelegramInfo.PhotoFileID = messages[i].Photo[len(messages[i].Photo)-1].FileID
		}
		if messages[i].Document != nil {
			picture.TelegramInfo.DocumentFileID = messages[i].Document.FileID
		}
	}

	for i, picture := range artwork.Pictures {
		info, err := storage.SavePicture(artwork, picture)
		if err != nil {
			Logger.Errorf("saving picture %s: %s", picture.Original, err)
			if telegram.Bot.DeleteMessages(&telego.DeleteMessagesParams{
				ChatID:     telegram.ChannelChatID,
				MessageIDs: telegram.GetMessageIDs(messages),
			}) != nil {
				Logger.Errorf("deleting messages: %s", err)
			}
			return fmt.Errorf("saving picture %s: %w", picture.Original, err)
		}
		artwork.Pictures[i].StorageInfo = info
	}

	_, err = service.CreateArtwork(ctx, artwork)
	if err != nil {
		go func() {
			if telegram.Bot.DeleteMessages(&telego.DeleteMessagesParams{
				ChatID:     telegram.ChannelChatID,
				MessageIDs: telegram.GetMessageIDs(messages),
			}) != nil {
				Logger.Errorf("deleting messages: %s", err)
			}
		}()
		return fmt.Errorf("error when creating artwork %s: %w", artwork.SourceURL, err)
	}
	return nil
}
