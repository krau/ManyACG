package bot

import (
	. "ManyACG-Bot/logger"
	"ManyACG-Bot/sources"
	"ManyACG-Bot/telegram"
	"os"
	"regexp"

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
	baseGroup.HandleMessageCtx(getArtworkInfo, telegohandler.Or(telegohandler.TextMatches(sources.AllSourceURLRegexp), telegohandler.CaptionMatches(sources.AllSourceURLRegexp)))

	adminHandlerGroup := botHandler.Group(
		telegohandler.Or(
			telegohandler.TextMatches(regexp.MustCompile(`^/(set_admin|del|delete|fetch)`)),
			telegohandler.CallbackDataPrefix("admin"),
		),
	)
	adminHandlerGroup.Use(adminCheck)

	adminHandlerGroup.HandleMessageCtx(setAdmin, telegohandler.CommandEqual("set_admin"))
	adminHandlerGroup.HandleMessageCtx(deletePicture, telegohandler.Or(telegohandler.CommandEqual("del"), telegohandler.CommandEqual("delete")))
	adminHandlerGroup.HandleMessageCtx(fetchArtwork, telegohandler.CommandEqual("fetch"))
	adminHandlerGroup.HandleCallbackQueryCtx(postArtwork, telegohandler.CallbackDataContains("post_artwork"))

	botHandler.Start()
}
