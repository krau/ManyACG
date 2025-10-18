package handlers

import (
	"errors"
	"strings"

	"github.com/krau/ManyACG/internal/infra/cache"
	"github.com/krau/ManyACG/internal/interface/telegram/handlers/utils"
	"github.com/krau/ManyACG/internal/interface/telegram/metautil"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/internal/shared/errs"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/samber/oops"
)

func Start(ctx *telegohandler.Context, message telego.Message) error {
	_, _, args := telegoutil.ParseCommand(message.Text)
	if len(args) > 0 {
		log.Debug("received start", "args", args)
		serv := service.FromContext(ctx)
		meta := metautil.MustFromContext(ctx)
		cmds := strings.Split(args[0], "_")
		action := cmds[0]
		switch action {
		case "file": // for compatibility we keep "file" action to get single picture by id
			pictureIDStr := args[0][5:]
			pictureID, err := objectuuid.FromObjectIDHex(pictureIDStr)
			if err != nil {
				utils.ReplyMessage(ctx, message, "无效的ID")
				return nil
			}
			picture, err := serv.GetPictureByID(ctx, pictureID)
			if err != nil {
				utils.ReplyMessage(ctx, message, "获取失败")
				return oops.Wrapf(err, "failed to get picture by id %s", pictureIDStr)
			}
			file, err := utils.GetPictureDocumentInputFile(ctx, serv, meta, picture.Artwork, picture)
			if err != nil {
				utils.ReplyMessage(ctx, message, "获取失败")
				return oops.Wrapf(err, "failed to get picture document input file by id %s", pictureIDStr)
			}
			defer file.Close()
			_, err = ctx.Bot().SendDocument(ctx, telegoutil.Document(message.Chat.ChatID(), file.Value))
			return err
		case "files":
			artworkIDStr := args[0][6:]
			artworkID, err := objectuuid.FromObjectIDHex(artworkIDStr)
			if err != nil {
				utils.ReplyMessage(ctx, message, "无效的ID")
				return nil
			}
			var artwork shared.ArtworkLike
			created, err := serv.GetArtworkByID(ctx, artworkID)
			if err != nil && !errors.Is(err, errs.ErrRecordNotFound) {
				utils.ReplyMessage(ctx, message, "获取失败")
				return oops.Wrapf(err, "failed to get artwork by id %s", artworkIDStr)
			}
			if created == nil {
				sourceUrl, ok := cache.Get[string](artworkIDStr)
				if !ok {
					utils.ReplyMessage(ctx, message, "获取失败")
					return oops.Errorf("failed to get string data by id: %s", artworkIDStr)
				}
				cached, err := serv.GetOrFetchCachedArtwork(ctx, sourceUrl)
				if err != nil {
					utils.ReplyMessage(ctx, message, "获取失败")
					return oops.Wrapf(err, "failed to get or fetch cached artwork by url: %s", sourceUrl)
				}
				artwork = cached
			} else {
				artwork = created
			}
			return getArtworkFiles(ctx, serv, meta, message, artwork)
		case "info":
			dataID := args[0][5:]
			sourceURL, ok := cache.Get[string](dataID)
			if !ok {
				utils.ReplyMessage(ctx, message, "获取失败")
				return oops.Errorf("failed to get string data by id: %s", dataID)
			}
			artwork, err := serv.GetOrFetchCachedArtwork(ctx, sourceURL)
			if err != nil {
				utils.ReplyMessage(ctx, message, "获取作品信息失败")
				return oops.Wrapf(err, "failed to get or fetch cached artwork by url: %s", sourceURL)
			}
			results, err := utils.SendArtworkMediaGroup(ctx, ctx.Bot(), serv, meta, message.Chat.ChatID(), artwork)
			if err != nil {
				utils.ReplyMessage(ctx, message, "发送作品信息失败")
				return oops.Wrapf(err, "failed to send artwork media group")
			}
			data := artwork.Artwork.Data()
			for _, res := range results {
				if res.UgoiraIndex >= 0 {
					if len(data.UgoiraMetas) <= res.UgoiraIndex {
						log.Warn("ugoira index out of range", "index", res.UgoiraIndex, "len", len(data.UgoiraMetas), "title", artwork.GetTitle(), "url", artwork.GetSourceURL())
						continue
					}
					data.UgoiraMetas[res.UgoiraIndex].TelegramInfo.SetFileID(meta.BotID(), shared.TelegramMediaTypeVideo, res.FileID)
				} else if res.PictureIndex >= 0 {
					if len(data.Pictures) <= res.PictureIndex {
						log.Warn("picture index out of range", "index", res.PictureIndex, "len", len(data.Pictures), "title", artwork.GetTitle(), "url", artwork.GetSourceURL())
						continue
					}
					data.Pictures[res.PictureIndex].TelegramInfo.SetFileID(meta.BotID(), shared.TelegramMediaTypePhoto, res.FileID)
				}
			}
			if err := serv.UpdateCachedArtwork(ctx, data); err != nil {
				return oops.Wrapf(err, "failed to update cached artwork after send media group")
			}
			return nil
		}
		return nil
	}
	return Help(ctx, message)
}
