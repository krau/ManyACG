package handlers

import (
	"github.com/krau/ManyACG/internal/app"
	"github.com/krau/ManyACG/internal/intf/telegram/handlers/filter"
	"github.com/krau/ManyACG/internal/intf/telegram/handlers/shared"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
)

type BotHandlers struct {
	channelChatID    telego.ChatID
	botUsername      string
	channelAvailable bool
	app              *app.Application
}

func New(meta *shared.HandlersMeta, app *app.Application) *BotHandlers {
	return &BotHandlers{
		channelChatID:    meta.ChannelChatID,
		botUsername:      meta.BotUsername,
		channelAvailable: meta.ChannelAvailable,
		app:              app,
	}
}

func (h *BotHandlers) Register(hg *telegohandler.HandlerGroup) {
	meta := &shared.HandlersMeta{
		ChannelChatID:    h.channelChatID,
		BotUsername:      h.botUsername,
		ChannelAvailable: h.channelAvailable,
	}
	hg.Use(func(ctx *telegohandler.Context, update telego.Update) error {
		ctx = ctx.WithValue(shared.HandlersMetaCtxKey{}, meta)
		return ctx.Next(update)
	})
	mg := hg.Group(telegohandler.AnyMessage(), filter.CommandToMe)
	mg.HandleMessage(h.Start, telegohandler.CommandEqual("start"))
	// mg.HandleMessage(GetPictureFile, telegohandler.Or(telegohandler.CommandEqual("file"), telegohandler.CommandEqual("files")))
	mg.HandleMessage(RandomPicture, telegohandler.Or(telegohandler.CommandEqual("setu"), telegohandler.CommandEqual("random")))
	mg.HandleMessage(h.Help, telegohandler.CommandEqual("help"))
	// mg.HandleMessage(SearchPicture, telegohandler.CommandEqual("search"))
	// mg.HandleMessage(GetArtworkInfoCommand, telegohandler.CommandEqual("info"))
	// mg.HandleMessage(CalculatePicture, telegohandler.CommandEqual("hash"))
	// mg.HandleMessage(GetStats, telegohandler.CommandEqual("stats"))
	// mg.HandleMessage(HybridSearchArtworks, telegohandler.CommandEqual("hybrid"))
	// mg.HandleMessage(SearchSimilarArtworks, telegohandler.CommandEqual("similar"))

	// Admin commands
	// mg.HandleMessage(SetAdmin, telegohandler.CommandEqual("set_admin"))
	// mg.HandleMessage(DeleteArtwork, telegohandler.Or(telegohandler.CommandEqual("delete"), telegohandler.CommandEqual("del")))
	// mg.HandleMessage(ToggleArtworkR18, telegohandler.CommandEqual("r18"))
	// mg.HandleMessage(SetArtworkTags, telegohandler.Or(telegohandler.CommandEqual("tags"), telegohandler.CommandEqual("addtags"), telegohandler.CommandEqual("deltags")))
	// mg.HandleMessage(EditArtworkTitle, telegohandler.CommandEqual("title"))
	// mg.HandleMessage(PostArtworkCommand, telegohandler.CommandEqual("post"))
	// mg.HandleMessage(RefreshArtwork, telegohandler.CommandEqual("refresh"))
	// mg.HandleMessage(AddTagAlias, telegohandler.CommandEqual("tagalias"))
	// mg.HandleMessage(DumpArtworkInfo, telegohandler.CommandEqual("dump"))
	// mg.HandleMessage(ReCaptionArtwork, telegohandler.CommandEqual("recaption"))
	// mg.HandleMessage(AutoTaggingArtwork, telegohandler.CommandEqual("autotag"))
	// mg.HandleMessage(ProcessPicturesHashAndSize, telegohandler.CommandEqual("process_pictures_hashsize"))
	// for migration
	// mg.HandleMessage(ProcessPicturesStorage, telegohandler.CommandEqual("process_pictures_storage"))
	// mg.HandleMessage(FixTwitterArtists, telegohandler.CommandEqual("fix_twitter_artists"))
	// mg.HandleMessage(AutoTagAllArtwork, telegohandler.CommandEqual("autotag_all"))

	// hg.HandleCallbackQuery(PostArtworkCallbackQuery, telegohandler.CallbackDataContains("post_artwork"))
	// hg.HandleCallbackQuery(SearchPictureCallbackQuery, telegohandler.CallbackDataPrefix("search_picture"))
	// hg.HandleCallbackQuery(ArtworkPreview, telegohandler.CallbackDataContains("artwork_preview"))
	// hg.HandleCallbackQuery(EditArtworkR18, telegohandler.CallbackDataPrefix("edit_artwork r18"))
	// hg.HandleCallbackQuery(DeleteArtworkCallbackQuery, telegohandler.CallbackDataPrefix("delete_artwork"))

	// hg.HandleInlineQuery(InlineQuery)
	// hg.HandleMessage(GetArtworkInfo, func(ctx context.Context, update telego.Update) bool {
	// 	message := update.Message
	// 	if message == nil {
	// 		return false
	// 	}
	// 	if update.Message.ViaBot != nil && update.Message.ViaBot.Username == h.botUsername {
	// 		return false
	// 	}
	// 	return utils.FindSourceURLForMessage(update.Message) != ""
	// })
}
