package bot

import (
	. "ManyACG/logger"
	"ManyACG/telegram"
	"os"

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

	baseGroup.HandleMessageCtx(start, telegohandler.CommandEqual("start"))
	baseGroup.HandleMessageCtx(getPictureFile, telegohandler.CommandEqual("file"))
	baseGroup.HandleMessageCtx(randomPicture, telegohandler.Or(telegohandler.CommandEqual("setu"), telegohandler.CommandEqual("random")))
	baseGroup.HandleMessageCtx(help, telegohandler.CommandEqual("help"))
	baseGroup.HandleMessageCtx(getArtworkInfo, sourceUrlMatches)
	baseGroup.HandleMessageCtx(searchPicture, telegohandler.CommandEqual("search"))
	baseGroup.HandleMessageCtx(setAdmin, telegohandler.CommandEqual("set_admin"))
	baseGroup.HandleMessageCtx(deletePicture, telegohandler.Or(telegohandler.CommandEqual("del"), telegohandler.CommandEqual("delete")))
	baseGroup.HandleMessageCtx(fetchArtwork, telegohandler.CommandEqual("fetch"))
	baseGroup.HandleMessageCtx(processPictures, telegohandler.CommandEqual("process_pictures"))
	baseGroup.HandleMessageCtx(setArtworkR18, telegohandler.CommandEqual("r18"))
	baseGroup.HandleMessageCtx(setArtworkTags, telegohandler.Or(telegohandler.CommandEqual("tags"), telegohandler.CommandEqual("addtags"), telegohandler.CommandEqual("deltags")))
	baseGroup.HandleCallbackQueryCtx(postArtwork, telegohandler.CallbackDataContains("post_artwork"))
	baseGroup.HandleInlineQueryCtx(inlineQuery)

	botHandler.Start()
}
