package handlers

import (
	"ManyACG/config"
	"ManyACG/telegram/utils"
	"context"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
)

var (
	ChannelChatID      telego.ChatID
	BotUsername        string
	IsChannelAvailable bool
)

func Init(channelChatID telego.ChatID, botUsername string) {
	ChannelChatID = channelChatID
	BotUsername = botUsername
	IsChannelAvailable = (ChannelChatID.ID != 0 || ChannelChatID.Username != "") && config.Cfg.Telegram.Channel
}

func RegisterHandlers(hg *telegohandler.HandlerGroup) {
	hg.HandleMessageCtx(Start, telegohandler.CommandEqual("start"), mentionIsBot)
	hg.HandleMessageCtx(GetPictureFile, telegohandler.Or(telegohandler.CommandEqual("file"), telegohandler.CommandEqual("files")), mentionIsBot)
	hg.HandleMessageCtx(RandomPicture, telegohandler.Or(telegohandler.CommandEqual("setu"), telegohandler.CommandEqual("random")), mentionIsBot)
	hg.HandleMessageCtx(Help, telegohandler.CommandEqual("help"), mentionIsBot)
	hg.HandleMessageCtx(SearchPicture, telegohandler.CommandEqual("search"), mentionIsBot)
	hg.HandleMessageCtx(CalculatePicture, telegohandler.CommandEqual("info"), mentionIsBot)
	hg.HandleMessageCtx(GetStats, telegohandler.CommandEqual("stats"), mentionIsBot)

	hg.HandleMessageCtx(SetAdmin, telegohandler.CommandEqual("set_admin"))
	hg.HandleMessageCtx(DeletePicture, telegohandler.Or(telegohandler.CommandEqual("del"), telegohandler.CommandEqual("delete")))
	hg.HandleMessageCtx(ProcessOldPictures, telegohandler.CommandEqual("process_pictures"))
	hg.HandleMessageCtx(SetArtworkR18, telegohandler.CommandEqual("r18"))
	hg.HandleMessageCtx(SetArtworkTags, telegohandler.Or(telegohandler.CommandEqual("tags"), telegohandler.CommandEqual("addtags"), telegohandler.CommandEqual("deltags")))
	hg.HandleMessageCtx(PostArtworkCommand, telegohandler.CommandEqual("post"))
	hg.HandleMessageCtx(BatchPostArtwork, telegohandler.CommandEqual("batch_post"))
	hg.HandleCallbackQueryCtx(PostArtworkCallbackQuery, telegohandler.CallbackDataContains("post_artwork"))
	hg.HandleCallbackQueryCtx(SearchPictureCallbackQuery, telegohandler.CallbackDataPrefix("search_picture"))
	hg.HandleCallbackQueryCtx(ArtworkPreview, telegohandler.CallbackDataContains("artwork_preview"))
	hg.HandleInlineQueryCtx(InlineQuery)

	sourceURLGroup := hg.Group(telegohandler.AnyMessage())
	sourceURLGroup.Use(func(bot *telego.Bot, update telego.Update, next telegohandler.Handler) {
		sourceURL := utils.FindSourceURLForMessage(update.Message)
		if sourceURL == "" {
			return
		}
		ctx := context.WithValue(update.Context(), "sourceURL", sourceURL)
		next(bot, update.WithContext(ctx))
	})

	sourceURLGroup.HandleMessageCtx(GetArtworkInfo)
}
