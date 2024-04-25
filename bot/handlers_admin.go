package bot

import (
	"ManyACG-Bot/config"
	"ManyACG-Bot/fetcher"
	"ManyACG-Bot/service"
	"ManyACG-Bot/sources"
	"ManyACG-Bot/storage"
	"ManyACG-Bot/telegram"
	"ManyACG-Bot/types"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	. "ManyACG-Bot/logger"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/mongo"
)

func setAdmin(ctx context.Context, bot *telego.Bot, message telego.Message) {
	userAdmin, err := service.GetAdminByUserID(ctx, message.From.ID)
	if err != nil {
		telegram.ReplyMessage(bot, message, "获取管理员信息失败: "+err.Error())
		return
	}
	if userAdmin != nil && !userAdmin.SuperAdmin {
		telegram.ReplyMessage(bot, message, "你没有权限设置管理员")
		return
	}
	var userID int64
	_, _, args := telegoutil.ParseCommand(message.Text)
	var permissions []types.Permission
	if message.ReplyToMessage != nil {
		if message.ReplyToMessage.SenderChat != nil {
			userID = message.ReplyToMessage.SenderChat.ID
		} else {
			userID = message.ReplyToMessage.From.ID
		}
		permissions = make([]types.Permission, len(args))
		unsupportedPermissions := make([]string, 0)
		for i, arg := range args {
			for _, p := range types.AllPermissions {
				if string(p) == arg {
					permissions[i] = p
					break
				}
			}
			if permissions[i] == "" {
				unsupportedPermissions = append(unsupportedPermissions, arg)
			}
		}
		if len(unsupportedPermissions) > 0 {
			text := fmt.Sprintf("权限不存在: %v\n支持的权限:\n", unsupportedPermissions)
			for _, p := range types.AllPermissions {
				text += string(p) + "\n"
			}
			telegram.ReplyMessage(bot, message, text)
			return
		}
	} else {
		if len(args) == 0 {
			telegram.ReplyMessage(bot, message, "请回复一条消息或提供用户ID, 并指定权限, 以空格分隔.\n若不提供权限则默认为所有权限")
			return
		}
		var err error
		userID, err = strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			telegram.ReplyMessage(bot, message, "请不要输入奇怪的东西")
			return
		}
		permissions = make([]types.Permission, len(args)-1)
		unsupportedPermissions := make([]string, 0)
		for i, arg := range args[1:] {
			for _, p := range types.AllPermissions {
				if string(p) == arg {
					permissions[i] = p
					break
				}
			}
			if permissions[i] == "" {
				unsupportedPermissions = append(unsupportedPermissions, arg)
			}
		}
		if len(unsupportedPermissions) > 0 {
			text := fmt.Sprintf("权限不存在: %v\n支持的权限:\n", unsupportedPermissions)
			for _, p := range types.AllPermissions {
				text += string(p) + "\n"
			}
			telegram.ReplyMessage(bot, message, text)
			return
		}
	}

	isAdmin, err := service.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			err := service.CreateOrUpdateAdmin(ctx, userID, permissions, message.From.ID, len(permissions) == 0)
			if err != nil {
				telegram.ReplyMessage(bot, message, "设置管理员失败: "+err.Error())
				return
			}
			telegram.ReplyMessage(bot, message, "设置管理员成功")
			return
		}
		telegram.ReplyMessage(bot, message, "获取管理员信息失败: "+err.Error())
		return
	}
	if isAdmin {
		if (len(args) == 0 && message.ReplyToMessage != nil) || (len(args) == 1 && message.ReplyToMessage == nil) {
			if err := service.DeleteAdmin(ctx, userID); err != nil {
				telegram.ReplyMessage(bot, message, "删除管理员失败: "+err.Error())
				return
			}
			telegram.ReplyMessage(bot, message, fmt.Sprintf("删除管理员成功: %d", userID))
			return
		}
		err := service.CreateOrUpdateAdmin(ctx, userID, permissions, message.From.ID, false)
		if err != nil {
			telegram.ReplyMessage(bot, message, "更新管理员失败: "+err.Error())
			return
		}
		telegram.ReplyMessage(bot, message, "更新管理员成功")
		return
	}
}

func deletePicture(ctx context.Context, bot *telego.Bot, message telego.Message) {
	var channelMessageID int
	cmd, _, args := telegoutil.ParseCommand(message.Text)
	if message.ReplyToMessage == nil {
		if len(args) == 0 {
			telegram.ReplyMessage(bot, message, "用法:\n使用 /del 回复一条频道的图片消息, 或者提供频道消息ID.\n若使用 /delete 则删除整个作品")
			return
		}
		var err error
		channelMessageID, err = strconv.Atoi(args[0])
		if err != nil {
			telegram.ReplyMessage(bot, message, "请不要输入奇怪的东西")
			return
		}
	} else {
		originChannel, ok := telegram.CheckTargetMessageIsChannelArtworkPost(ctx, bot, message)
		if !ok {
			telegram.ReplyMessage(bot, message, "用法:\n使用 /del 回复一条频道的图片消息, 或者提供频道消息ID.\n若使用 /delete 则删除整个作品")
			return
		}
		channelMessageID = originChannel.MessageID
	}
	if cmd == "del" {
		if !service.CheckAdminPermission(ctx, message.From.ID, types.DeletePicture) {
			telegram.ReplyMessage(bot, message, "你没有删除图片的权限")
			return
		}
		picture, err := service.GetPictureByMessageID(ctx, channelMessageID)
		if err != nil {
			telegram.ReplyMessage(bot, message, "获取图片信息失败: "+err.Error())
			return
		}
		if err := service.DeletePictureByMessageID(ctx, channelMessageID); err != nil {
			telegram.ReplyMessage(bot, message, "从数据库中删除失败: "+err.Error())
			return
		}
		go telegram.ReplyMessage(bot, message, fmt.Sprintf("删除成功: %d", channelMessageID))
		go bot.DeleteMessage(telegoutil.Delete(telegram.ChannelChatID, channelMessageID))

		if err := storage.GetStorage().DeletePicture(picture.StorageInfo); err != nil {
			Logger.Warnf("删除图片失败: %s", err)
			bot.SendMessage(telegoutil.Message(telegoutil.ID(message.From.ID), "删除图片失败: "+err.Error()))
		}
		return
	}
	if !service.CheckAdminPermission(ctx, message.From.ID, types.DeleteArtwork) {
		telegram.ReplyMessage(bot, message, "你没有删除作品的权限")
		return
	}
	artwork, err := service.GetArtworkByMessageID(ctx, channelMessageID)
	if err != nil {
		telegram.ReplyMessage(bot, message, "获取作品信息失败: "+err.Error())
		return
	}
	if err := service.DeleteArtworkByURL(ctx, artwork.SourceURL); err != nil {
		telegram.ReplyMessage(bot, message, "从数据库中删除失败: "+err.Error())
		return
	}
	go telegram.ReplyMessage(bot, message, fmt.Sprintf("删除成功: %d", channelMessageID))
	artworkMessageIDs := make([]int, len(artwork.Pictures))
	for i, picture := range artwork.Pictures {
		artworkMessageIDs[i] = picture.TelegramInfo.MessageID
	}
	go bot.DeleteMessages(&telego.DeleteMessagesParams{
		ChatID:     telegram.ChannelChatID,
		MessageIDs: artworkMessageIDs,
	})

	for _, picture := range artwork.Pictures {
		if err := storage.GetStorage().DeletePicture(picture.StorageInfo); err != nil {
			Logger.Warnf("删除图片失败: %s", err)
			bot.SendMessage(telegoutil.Message(telegoutil.ID(message.From.ID), "删除图片失败: "+err.Error()))
		}
	}
}

func fetchArtwork(ctx context.Context, bot *telego.Bot, message telego.Message) {
	if !service.CheckAdminPermission(ctx, message.From.ID, types.FetchArtwork) {
		telegram.ReplyMessage(bot, message, "你没有拉取作品的权限")
		return
	}

	go fetcher.FetchOnce(context.TODO(), config.Cfg.Fetcher.Limit)
	telegram.ReplyMessage(bot, message, "开始拉取作品了")
}

func getArtworkInfoForAdmin(ctx context.Context, bot *telego.Bot, message telego.Message) {
	if !service.CheckAdminPermission(ctx, message.From.ID, types.GetArtworkInfo) {
		telegram.ReplyMessage(bot, message, "你没有获取作品信息的权限")
		return
	}
	sourceURL := sources.MatchSourceURL(message.Text)
	if sourceURL == "" {
		return
	}
	var waitMessageID int

	go func() {
		msg, err := telegram.ReplyMessage(bot, message, "正在获取作品信息...")
		if err != nil {
			Logger.Warnf("发送消息失败: %s", err)
			return
		}
		waitMessageID = msg.MessageID
	}()

	defer func() {
		if r := recover(); r != nil {
			Logger.Errorf("panic: %v", r)
		}
		time.Sleep(1 * time.Second)
		if waitMessageID != 0 {
			go bot.DeleteMessage(telegoutil.Delete(message.Chat.ChatID(), waitMessageID))
		}
	}()

	isAlreadyPosted := true
	artwork, err := service.GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		isAlreadyPosted = false
		artwork, err = sources.GetArtworkInfo(sourceURL)
	}
	if err != nil {
		telegram.ReplyMessage(bot, message, "获取作品信息失败: "+err.Error())
		return
	}

	var inputFile telego.InputFile
	if artwork.Pictures[0].TelegramInfo == nil || artwork.Pictures[0].TelegramInfo.PhotoFileID == "" {
		photoURL := artwork.Pictures[0].Original
		if artwork.SourceType == types.SourceTypePixiv {
			photoURL = sources.GetPixivRegularURL(photoURL)
		}
		inputFile = telegoutil.FileFromURL(photoURL)
	} else {
		inputFile = telegoutil.FileFromID(artwork.Pictures[0].TelegramInfo.PhotoFileID)
	}

	photo := telegoutil.Photo(message.Chat.ChatID(), inputFile).
		WithReplyParameters(&telego.ReplyParameters{MessageID: message.MessageID}).
		WithParseMode(telego.ModeMarkdownV2)

	deletedModel, _ := service.GetDeletedByURL(ctx, sourceURL)
	artworkInfoCaption := telegram.GetArtworkMarkdownCaption(artwork)
	if deletedModel != nil {
		photo.WithCaption(artworkInfoCaption + telegram.EscapeMarkdown(fmt.Sprintf("\n\n这是一个在 %s 删除的作品\n\n"+
			"如果发布则会取消删除", deletedModel.DeletedAt.Time().Format("2006-01-02 15:04:05"))))
	} else {
		photo.WithCaption(telegram.GetArtworkMarkdownCaption(artwork))
	}
	if isAlreadyPosted {
		photo.WithReplyMarkup(telegoutil.InlineKeyboard(
			[]telego.InlineKeyboardButton{
				telegoutil.InlineKeyboardButton("来源").WithURL(telegram.GetArtworkPostMessageURL(artwork.Pictures[0].TelegramInfo.MessageID)),
				telegoutil.InlineKeyboardButton("原图").WithURL(telegram.GetDeepLinkForFile(artwork.Pictures[0].TelegramInfo.MessageID)),
			},
		))
	} else {
		id, err := service.CreateCallbackData(ctx, artwork.SourceURL)
		if err != nil {
			telegram.ReplyMessage(bot, message, "创建回调数据失败: "+err.Error())
			return
		}
		photo.WithReplyMarkup(telegoutil.InlineKeyboard(
			[]telego.InlineKeyboardButton{
				telegoutil.InlineKeyboardButton("发布到频道").WithCallbackData("admin post_artwork " + id),
			},
		))
	}
	if artwork.R18 {
		photo.WithHasSpoiler()
	}
	_, err = bot.SendPhoto(photo)
	if err != nil {
		telegram.ReplyMessage(bot, message, "发送图片失败: "+err.Error())
		return
	}
}

func postArtwork(ctx context.Context, bot *telego.Bot, query telego.CallbackQuery) {
	if !service.CheckAdminPermission(ctx, query.From.ID, types.PostArtwork) {
		bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "你没有发布作品的权限",
			ShowAlert:       true,
			CacheTime:       60,
		})
		return
	}
	dataID := strings.Split(query.Data, " ")[2]
	sourceURL, err := service.GetCallbackDataByID(ctx, dataID)
	if err != nil {
		bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "获取回调数据失败" + err.Error(),
			ShowAlert:       true,
			CacheTime:       60,
		})
		return
	}
	Logger.Infof("posting artwork: %s", sourceURL)
	artwork, err := sources.GetArtworkInfo(sourceURL)
	if err != nil {
		bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "获取作品信息失败" + err.Error(),
			ShowAlert:       true,
			CacheTime:       60,
		})
		return
	}
	go bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
		CallbackQueryID: query.ID,
		Text:            "正在发布, 请不要重复点击...",
		CacheTime:       120,
	})
	if service.CheckDeletedByURL(ctx, sourceURL) {
		if err := service.DeleteDeletedByURL(ctx, sourceURL); err != nil {
			Logger.Errorf("删除删除记录失败: %s", err)
			go bot.EditMessageCaption(&telego.EditMessageCaptionParams{
				ChatID:    telegoutil.ID(query.Message.GetChat().ID),
				MessageID: query.Message.GetMessageID(),
				Caption:   "取消删除记录失败: " + err.Error(),
			})
			return
		}
	}
	if err := fetcher.PostAndCreateArtwork(ctx, artwork, bot, storage.GetStorage()); err != nil {
		Logger.Errorf("发布失败: %s", err)
		go bot.EditMessageCaption(&telego.EditMessageCaptionParams{
			ChatID:    telegoutil.ID(query.Message.GetChat().ID),
			MessageID: query.Message.GetMessageID(),
			Caption:   "发布失败: " + err.Error() + "\n\n" + time.Now().Format("2006-01-02 15:04:05"),
		})
		return
	}
	artwork, err = service.GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		Logger.Errorf("获取发布后的作品信息失败: %s", err)
		go bot.EditMessageCaption(&telego.EditMessageCaptionParams{
			ChatID:      telegoutil.ID(query.Message.GetChat().ID),
			MessageID:   query.Message.GetMessageID(),
			Caption:     "发布成功, 但获取作品信息失败: " + err.Error(),
			ReplyMarkup: nil,
		})
		return
	}
	go bot.EditMessageCaption(&telego.EditMessageCaptionParams{
		ChatID:    telegoutil.ID(query.Message.GetChat().ID),
		MessageID: query.Message.GetMessageID(),
		Caption:   "发布成功: " + artwork.Title + "\n\n发布时间: " + artwork.CreatedAt.Format("2006-01-02 15:04:05"),
		ReplyMarkup: telegoutil.InlineKeyboard(
			[]telego.InlineKeyboardButton{
				telegoutil.InlineKeyboardButton("去查看").WithURL(telegram.GetArtworkPostMessageURL(artwork.Pictures[0].TelegramInfo.MessageID)),
			},
		),
	})
}
