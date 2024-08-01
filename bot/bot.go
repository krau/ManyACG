package bot

import (
	. "ManyACG/logger"
	"ManyACG/telegram"
	"os"

	"ManyACG/bot/handlers"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
)

func RunPolling() {
	if telegram.Bot == nil {
		telegram.InitBot()
	}
	Logger.Info("Start polling")
	updates, err := telegram.Bot.UpdatesViaLongPolling(&telego.GetUpdatesParams{
		Offset: -1,
		AllowedUpdates: []string{
			telego.MessageUpdates,
			telego.ChannelPostUpdates,
			telego.CallbackQueryUpdates,
			telego.InlineQueryUpdates,
		},
	})
	if err != nil {
		Logger.Fatalf("Error when getting updates: %s", err)
		os.Exit(1)
	}

	botHandler, err := telegohandler.NewBotHandler(telegram.Bot, updates)
	if err != nil {
		Logger.Fatalf("Error when creating bot handler: %s", err)
		os.Exit(1)
	}
	defer botHandler.Stop()
	defer telegram.Bot.StopLongPolling()

	botHandler.Use(telegohandler.PanicRecovery())
	botHandler.Use(messageLogger)
	baseGroup := botHandler.BaseGroup()

	baseGroup.HandleMessageCtx(handlers.Start, telegohandler.CommandEqual("start"), mentionIsBot)
	baseGroup.HandleMessageCtx(handlers.GetPictureFile, telegohandler.Or(telegohandler.CommandEqual("file"), telegohandler.CommandEqual("files")), mentionIsBot)
	baseGroup.HandleMessageCtx(handlers.RandomPicture, telegohandler.Or(telegohandler.CommandEqual("setu"), telegohandler.CommandEqual("random")), mentionIsBot)
	baseGroup.HandleMessageCtx(handlers.Help, telegohandler.CommandEqual("help"), mentionIsBot)
	baseGroup.HandleMessageCtx(handlers.SearchPicture, telegohandler.CommandEqual("search"), mentionIsBot)
	baseGroup.HandleMessageCtx(handlers.CalculatePicture, telegohandler.CommandEqual("info"), mentionIsBot)
	baseGroup.HandleMessageCtx(handlers.GetStats, telegohandler.CommandEqual("stats"), mentionIsBot)

	baseGroup.HandleMessageCtx(handlers.SetAdmin, telegohandler.CommandEqual("set_admin"))
	baseGroup.HandleMessageCtx(handlers.DeletePicture, telegohandler.Or(telegohandler.CommandEqual("del"), telegohandler.CommandEqual("delete")))
	baseGroup.HandleMessageCtx(handlers.FetchArtwork, telegohandler.CommandEqual("fetch"))
	baseGroup.HandleMessageCtx(handlers.ProcessOldPictures, telegohandler.CommandEqual("process_pictures"))
	baseGroup.HandleMessageCtx(handlers.SetArtworkR18, telegohandler.CommandEqual("r18"))
	baseGroup.HandleMessageCtx(handlers.SetArtworkTags, telegohandler.Or(telegohandler.CommandEqual("tags"), telegohandler.CommandEqual("addtags"), telegohandler.CommandEqual("deltags")))
	baseGroup.HandleMessageCtx(handlers.PostArtworkCommand, telegohandler.CommandEqual("post"))
	baseGroup.HandleMessageCtx(handlers.BatchPostArtwork, telegohandler.CommandEqual("batch_post"))
	baseGroup.HandleCallbackQueryCtx(handlers.PostArtworkCallbackQuery, telegohandler.CallbackDataContains("post_artwork"))
	baseGroup.HandleCallbackQueryCtx(handlers.ArtworkPreview, telegohandler.CallbackDataContains("artwork_preview"))
	baseGroup.HandleMessageCtx(handlers.GetArtworkInfo, func(update telego.Update) bool {
		return telegram.FindSourceURLForMessage(update.Message) != ""
	})
	baseGroup.HandleInlineQueryCtx(handlers.InlineQuery)

	botHandler.Start()
}
