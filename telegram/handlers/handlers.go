package handlers

import (
	"ManyACG/config"
	"ManyACG/telegram/utils"

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
	hg.HandleMessageCtx(DeleteArtwork, telegohandler.CommandEqual("delete"))
	hg.HandleMessageCtx(ProcessPicturesHashAndSize, telegohandler.CommandEqual("process_pictures_hashsize"))
	hg.HandleMessageCtx(ProcessPicturesStorage, telegohandler.CommandEqual("process_pictures_storage"))
	hg.HandleMessageCtx(FixTwitterArtists, telegohandler.CommandEqual("fix_twitter_artists"))

	hg.HandleMessageCtx(SetArtworkR18, telegohandler.CommandEqual("r18"))
	hg.HandleMessageCtx(SetArtworkTags, telegohandler.Or(telegohandler.CommandEqual("tags"), telegohandler.CommandEqual("addtags"), telegohandler.CommandEqual("deltags")))
	hg.HandleMessageCtx(EditArtworkTitle, telegohandler.CommandEqual("title"))
	hg.HandleMessageCtx(PostArtworkCommand, telegohandler.CommandEqual("post"))
	// hg.HandleMessageCtx(BatchPostArtwork, telegohandler.CommandEqual("batch_post")) // TODO: 兼容无频道模式
	hg.HandleCallbackQueryCtx(PostArtworkCallbackQuery, telegohandler.CallbackDataContains("post_artwork"))
	hg.HandleCallbackQueryCtx(SearchPictureCallbackQuery, telegohandler.CallbackDataPrefix("search_picture"))
	hg.HandleCallbackQueryCtx(ArtworkPreview, telegohandler.CallbackDataContains("artwork_preview"))
	hg.HandleCallbackQueryCtx(EditArtworkR18, telegohandler.CallbackDataPrefix("edit_artwork r18"))
	hg.HandleCallbackQueryCtx(DeleteArtworkCallbackQuery, telegohandler.CallbackDataPrefix("delete_artwork"))

	hg.HandleInlineQueryCtx(InlineQuery)
	hg.HandleMessageCtx(GetArtworkInfo, func(update telego.Update) bool {
		if update.Message.ViaBot != nil && update.Message.ViaBot.Username == BotUsername {
			return false
		}
		return utils.FindSourceURLForMessage(update.Message) != ""
	})
}
