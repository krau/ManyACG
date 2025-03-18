package handlers

import (
	"context"

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
	mg := hg.Group(telegohandler.AnyMessage(), commandToMe)
	mg.HandleMessage(Start, telegohandler.CommandEqual("start"))
	mg.HandleMessage(GetPictureFile, telegohandler.Or(telegohandler.CommandEqual("file"), telegohandler.CommandEqual("files")))
	mg.HandleMessage(RandomPicture, telegohandler.Or(telegohandler.CommandEqual("setu"), telegohandler.CommandEqual("random")))
	mg.HandleMessage(Help, telegohandler.CommandEqual("help"))
	mg.HandleMessage(SearchPicture, telegohandler.CommandEqual("search"))
	mg.HandleMessage(GetArtworkInfoCommand, telegohandler.CommandEqual("info"))
	mg.HandleMessage(CalculatePicture, telegohandler.CommandEqual("hash"))
	mg.HandleMessage(GetStats, telegohandler.CommandEqual("stats"))
	mg.HandleMessage(HybridSearchArtworks, telegohandler.CommandEqual("hybrid"))
	mg.HandleMessage(SearchSimilarArtworks, telegohandler.CommandEqual("similar"))

	// Admin commands
	mg.HandleMessage(SetAdmin, telegohandler.CommandEqual("set_admin"))
	mg.HandleMessage(DeleteArtwork, telegohandler.Or(telegohandler.CommandEqual("delete"), telegohandler.CommandEqual("del")))
	mg.HandleMessage(ToggleArtworkR18, telegohandler.CommandEqual("r18"))
	mg.HandleMessage(SetArtworkTags, telegohandler.Or(telegohandler.CommandEqual("tags"), telegohandler.CommandEqual("addtags"), telegohandler.CommandEqual("deltags")))
	mg.HandleMessage(EditArtworkTitle, telegohandler.CommandEqual("title"))
	mg.HandleMessage(PostArtworkCommand, telegohandler.CommandEqual("post"))
	mg.HandleMessage(RefreshArtwork, telegohandler.CommandEqual("refresh"))
	mg.HandleMessage(AddTagAlias, telegohandler.CommandEqual("tagalias"))
	mg.HandleMessage(DumpArtworkInfo, telegohandler.CommandEqual("dump"))
	mg.HandleMessage(ReCaptionArtwork, telegohandler.CommandEqual("recaption"))
	mg.HandleMessage(AutoTaggingArtwork, telegohandler.CommandEqual("autotag"))
	mg.HandleMessage(ProcessPicturesHashAndSize, telegohandler.CommandEqual("process_pictures_hashsize"))
	mg.HandleMessage(ProcessPicturesStorage, telegohandler.CommandEqual("process_pictures_storage"))
	mg.HandleMessage(FixTwitterArtists, telegohandler.CommandEqual("fix_twitter_artists"))
	mg.HandleMessage(AutoTagAllArtwork, telegohandler.CommandEqual("autotag_all"))

	hg.HandleCallbackQuery(PostArtworkCallbackQuery, telegohandler.CallbackDataContains("post_artwork"))
	hg.HandleCallbackQuery(SearchPictureCallbackQuery, telegohandler.CallbackDataPrefix("search_picture"))
	hg.HandleCallbackQuery(ArtworkPreview, telegohandler.CallbackDataContains("artwork_preview"))
	hg.HandleCallbackQuery(EditArtworkR18, telegohandler.CallbackDataPrefix("edit_artwork r18"))
	hg.HandleCallbackQuery(DeleteArtworkCallbackQuery, telegohandler.CallbackDataPrefix("delete_artwork"))

	hg.HandleInlineQuery(InlineQuery)
	hg.HandleMessage(GetArtworkInfo, func(ctx context.Context, update telego.Update) bool {
		if update.Message.ViaBot != nil && update.Message.ViaBot.Username == BotUsername {
			return false
		}
		return utils.FindSourceURLForMessage(update.Message) != ""
	})
}
