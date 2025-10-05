package handlers

import (
	"github.com/krau/ManyACG/internal/interface/telegram/handlers/filter"
	"github.com/krau/ManyACG/internal/interface/telegram/handlers/utils"
	"github.com/krau/ManyACG/internal/interface/telegram/metautil"
	"github.com/krau/ManyACG/service"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
)

type HandlerManager struct {
	*metautil.MetaData
	*service.Service
}

func New(meta *metautil.MetaData, serv *service.Service) *HandlerManager {
	return &HandlerManager{
		MetaData: meta,
		Service:  serv,
	}
}

func (m HandlerManager) Register(hg *telegohandler.HandlerGroup) {
	hg.Handle(func(ctx *telegohandler.Context, update telego.Update) error {
		servCtx := service.WithContext(ctx, m.Service)
		metaCtx := metautil.WithContext(servCtx, m.MetaData)
		ctx = ctx.WithContext(metaCtx)
		return ctx.Next(update)
	})
	mg := hg.Group(telegohandler.AnyMessage(), filter.CommandToMe)
	mg.HandleMessage(Start, telegohandler.CommandEqual("start"))
	mg.HandleMessage(GetPictureFile, telegohandler.Or(telegohandler.CommandEqual("file"), telegohandler.CommandEqual("files")))
	// mg.HandleMessage(RandomPicture, telegohandler.Or(telegohandler.CommandEqual("setu"), telegohandler.CommandEqual("random")))
	mg.HandleMessage(Help, telegohandler.CommandEqual("help"))
	// mg.HandleMessage(SearchPicture, telegohandler.CommandEqual("search"))
	mg.HandleMessage(GetArtworkInfoCommand, telegohandler.CommandEqual("info"))
	// mg.HandleMessage(CalculatePicture, telegohandler.CommandEqual("hash"))
	// mg.HandleMessage(GetStats, telegohandler.CommandEqual("stats"))
	// mg.HandleMessage(HybridSearchArtworks, telegohandler.CommandEqual("hybrid"))
	// mg.HandleMessage(SearchSimilarArtworks, telegohandler.CommandEqual("similar"))

	// Admin commands
	// mg.HandleMessage(SetAdmin, telegohandler.CommandEqual("set_admin"))
	mg.HandleMessage(DeleteArtwork, telegohandler.Or(telegohandler.CommandEqual("delete"), telegohandler.CommandEqual("del")))
	mg.HandleMessage(ToggleArtworkR18, telegohandler.CommandEqual("r18"))
	mg.HandleMessage(SetArtworkTags, telegohandler.Or(telegohandler.CommandEqual("tags"), telegohandler.CommandEqual("addtags"), telegohandler.CommandEqual("deltags")))
	mg.HandleMessage(EditArtworkTitle, telegohandler.CommandEqual("title"))
	mg.HandleMessage(PostArtworkCommand, telegohandler.CommandEqual("post"))
	mg.HandleMessage(RefreshArtwork, telegohandler.CommandEqual("refresh"))
	// mg.HandleMessage(AddTagAlias, telegohandler.CommandEqual("tagalias"))
	mg.HandleMessage(DumpArtworkInfo, telegohandler.CommandEqual("dump"))
	mg.HandleMessage(ReCaptionArtwork, telegohandler.CommandEqual("recaption"))
	// mg.HandleMessage(AutoTaggingArtwork, telegohandler.CommandEqual("autotag"))
	// mg.HandleMessage(ProcessPicturesHashAndSize, telegohandler.CommandEqual("process_pictures_hashsize"))
	// for migration
	// mg.HandleMessage(ProcessPicturesStorage, telegohandler.CommandEqual("process_pictures_storage"))
	// mg.HandleMessage(FixTwitterArtists, telegohandler.CommandEqual("fix_twitter_artists"))
	// mg.HandleMessage(AutoTagAllArtwork, telegohandler.CommandEqual("autotag_all"))

	hg.HandleCallbackQuery(PostArtworkCallbackQuery, telegohandler.CallbackDataContains("post_artwork"))
	// hg.HandleCallbackQuery(SearchPictureCallbackQuery, telegohandler.CallbackDataPrefix("search_picture"))
	hg.HandleCallbackQuery(ArtworkPreview, telegohandler.CallbackDataContains("artwork_preview"))
	hg.HandleCallbackQuery(EditArtworkR18, telegohandler.CallbackDataPrefix("edit_artwork r18"))
	hg.HandleCallbackQuery(DeleteArtworkCallbackQuery, telegohandler.CallbackDataPrefix("delete_artwork"))

	hg.HandleInlineQuery(InlineQuery)
	hg.Use(func(ctx *telegohandler.Context, update telego.Update) error {
		msg := update.Message
		if msg == nil {
			return ctx.Err()
		}
		if update.Message.ViaBot != nil && update.Message.ViaBot.Username == m.BotUsername {
			return ctx.Err()
		}
		if url := utils.FindSourceURLInMessage(m.Service, update.Message); url != "" {
			ctx = ctx.WithValue("source_url", url)
			return ctx.Next(update)
		}
		return ctx.Err()
	})
	hg.HandleMessage(GetArtworkInfo)
}
