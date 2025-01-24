package handlers

import (
	"github.com/krau/ManyACG/config"
	"github.com/krau/ManyACG/telegram/utils"

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
	mg := hg.Group(telegohandler.AnyMessage(), mentionIsBot)
	mg.HandleMessageCtx(Start, telegohandler.CommandEqual("start"))
	mg.HandleMessageCtx(GetPictureFile, telegohandler.Or(telegohandler.CommandEqual("file"), telegohandler.CommandEqual("files")))
	mg.HandleMessageCtx(RandomPicture, telegohandler.Or(telegohandler.CommandEqual("setu"), telegohandler.CommandEqual("random")))
	mg.HandleMessageCtx(Help, telegohandler.CommandEqual("help"))
	mg.HandleMessageCtx(SearchPicture, telegohandler.CommandEqual("search"))
	mg.HandleMessageCtx(GetArtworkInfoCommand, telegohandler.CommandEqual("info"))
	mg.HandleMessageCtx(CalculatePicture, telegohandler.CommandEqual("hash"))
	mg.HandleMessageCtx(GetStats, telegohandler.CommandEqual("stats"))
	mg.HandleMessageCtx(HybridSearchArtworks, telegohandler.CommandEqual("query"))

	mg.HandleMessageCtx(SetAdmin, telegohandler.CommandEqual("set_admin"))
	mg.HandleMessageCtx(DeleteArtwork, telegohandler.Or(telegohandler.CommandEqual("delete"), telegohandler.CommandEqual("del")))
	mg.HandleMessageCtx(ProcessPicturesHashAndSize, telegohandler.CommandEqual("process_pictures_hashsize"))
	mg.HandleMessageCtx(ProcessPicturesStorage, telegohandler.CommandEqual("process_pictures_storage"))
	mg.HandleMessageCtx(FixTwitterArtists, telegohandler.CommandEqual("fix_twitter_artists"))
	mg.HandleMessageCtx(ToggleArtworkR18, telegohandler.CommandEqual("r18"))
	mg.HandleMessageCtx(SetArtworkTags, telegohandler.Or(telegohandler.CommandEqual("tags"), telegohandler.CommandEqual("addtags"), telegohandler.CommandEqual("deltags")))
	mg.HandleMessageCtx(EditArtworkTitle, telegohandler.CommandEqual("title"))
	mg.HandleMessageCtx(PostArtworkCommand, telegohandler.CommandEqual("post"))
	mg.HandleMessageCtx(RefreshArtwork, telegohandler.CommandEqual("refresh"))
	// hg.HandleMessageCtx(BatchPostArtwork, telegohandler.CommandEqual("batch_post")) // TODO: 兼容无频道模式
	mg.HandleMessageCtx(AddTagAlias, telegohandler.CommandEqual("tagalias"))
	mg.HandleMessageCtx(DumpArtworkInfo, telegohandler.CommandEqual("dump"))
	mg.HandleMessageCtx(ReCaptionArtwork, telegohandler.CommandEqual("recaption"))

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
