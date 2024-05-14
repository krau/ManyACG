package fetcher

import (
	"ManyACG/errors"
	"ManyACG/service"
	"ManyACG/storage"
	"ManyACG/telegram"
	"ManyACG/types"
	"context"
	es "errors"
	"fmt"
	"sync"

	. "ManyACG/logger"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/mongo"
)

func PostAndCreateArtwork(ctx context.Context, artwork *types.Artwork, bot *telego.Bot, storage storage.Storage, fromID int64) error {
	artworkInDB, err := service.GetArtworkByURL(ctx, artwork.SourceURL)
	if err == nil && artworkInDB != nil {
		Logger.Debugf("Artwork %s already exists", artwork.Title)
		return errors.ErrArtworkAlreadyExist
	}
	if err != nil && !es.Is(err, mongo.ErrNoDocuments) {
		return err
	}
	if service.CheckDeletedByURL(ctx, artwork.SourceURL) {
		Logger.Debugf("Artwork %s is deleted", artwork.Title)
		return errors.ErrArtworkDeleted
	}
	for i, picture := range artwork.Pictures {
		info, err := storage.SavePicture(artwork, picture)
		if err != nil {
			Logger.Errorf("saving picture %s: %s", picture.Original, err)
			return fmt.Errorf("saving picture %s: %w", picture.Original, err)
		}
		artwork.Pictures[i].StorageInfo = info
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
	go afterCreate(context.TODO(), artwork, bot, storage, fromID)
	return nil
}

func afterCreate(ctx context.Context, artwork *types.Artwork, bot *telego.Bot, _ storage.Storage, fromID int64) {
	var wg sync.WaitGroup
	for _, picture := range artwork.Pictures {
		wg.Add(1)
		go func(picture *types.Picture) {
			defer wg.Done()
			if err := service.ProcessPictureAndUpdate(ctx, picture); err != nil {
				Logger.Warnf("error when processing %d of artwork %s: %s", picture.Index, artwork.Title, err)
			}
		}(picture)
	}
	wg.Wait()

	sendNotify := fromID != 0 && bot != nil
	artworkID, err := service.GetArtworkIDByPicture(ctx, artwork.Pictures[0])
	artworkTitleMarkdown := telegram.EscapeMarkdown(artwork.Title)
	if err != nil {
		Logger.Errorf("error when getting artwork by URL: %s", err)
		if sendNotify {
			bot.SendMessage(telegoutil.Messagef(telegoutil.ID(fromID),
				"刚刚发布的作品 [%s](%s) 后续处理失败\\: \n无法获取作品信息\\: %s", artworkTitleMarkdown, telegram.GetArtworkPostMessageURL(artwork.Pictures[0].TelegramInfo.MessageID), err).
				WithParseMode(telego.ModeMarkdownV2))
		}
		return
	}

	for _, picture := range artwork.Pictures {
		picture, err = service.GetPictureByMessageID(ctx, picture.TelegramInfo.MessageID)
		if err != nil {
			Logger.Warnf("error when getting picture by message ID: %s", err)
			continue
		}
		resPictures, err := service.GetPicturesByHash(ctx, picture.Hash)
		if err != nil {
			Logger.Warnf("error when getting pictures by hash: %s", err)
			continue
		}
		similarPictures := make([]*types.Picture, 0)
		for _, resPicture := range resPictures {
			resArtworkID, err := service.GetArtworkIDByPicture(ctx, resPicture)
			if err != nil {
				Logger.Warnf("error when getting artwork ID by picture: %s", err)
				continue
			}
			if resArtworkID == artworkID {
				continue
			}
			similarPictures = append(similarPictures, resPicture)
		}
		if len(similarPictures) == 0 {
			continue
		}
		Logger.Noticef("Found %d similar pictures for %s", len(similarPictures), picture.Original)
		if !sendNotify {
			continue
		}

		text := fmt.Sprintf("*刚刚发布的作品 [%s](%s) 中第 %d 张图片搜索到有%d个相似图片*\n",
			artworkTitleMarkdown,
			telegram.EscapeMarkdown(telegram.GetArtworkPostMessageURL(picture.TelegramInfo.MessageID)),
			picture.Index+1,
			len(similarPictures))
		text += telegram.EscapeMarkdown(fmt.Sprintf("该图像模糊度: %.2f\n搜索到的相似图片列表:\n\n", picture.BlurScore))
		for _, similarPicture := range similarPictures {
			artworkOfSimilarPicture, err := service.GetArtworkByMessageID(ctx, similarPicture.TelegramInfo.MessageID)
			if err != nil {
				text += telegram.EscapeMarkdown(fmt.Sprintf("%s 模糊度: %.2f\n\n", telegram.GetArtworkPostMessageURL(picture.TelegramInfo.MessageID), similarPicture.BlurScore))
				continue
			}
			text += fmt.Sprintf("[%s\\_%d](%s)  ",
				telegram.EscapeMarkdown(artworkOfSimilarPicture.Title),
				similarPicture.Index+1,
				telegram.EscapeMarkdown(telegram.GetArtworkPostMessageURL(similarPicture.TelegramInfo.MessageID)))
			text += telegram.EscapeMarkdown(fmt.Sprintf("模糊度: %.2f\n\n", similarPicture.BlurScore))
		}
		text += "_模糊度使用原图文件计算得出, 越小图像质量越好_"
		_, err = bot.SendMessage(telegoutil.Messagef(telegoutil.ID(fromID), text).WithParseMode(telego.ModeMarkdownV2))
		if err != nil {
			Logger.Errorf("error when sending similar pictures: %s", err)
		}
	}
}
