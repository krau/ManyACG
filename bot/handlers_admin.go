package bot

import (
	"ManyACG/common"
	"ManyACG/config"
	"ManyACG/fetcher"
	"ManyACG/service"
	"ManyACG/sources"
	"ManyACG/storage"
	"ManyACG/telegram"
	"ManyACG/types"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	. "ManyACG/logger"

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
	var userIDStr string
	_, _, args := telegoutil.ParseCommand(message.Text)
	var supportedPermissionsText string
	for _, p := range types.AllPermissions {
		supportedPermissionsText += "`" + string(p) + "`" + "\n"
	}
	if message.ReplyToMessage == nil {
		if len(args) == 0 {
			telegram.ReplyMessageWithMarkdown(
				bot, message,
				fmt.Sprintf("请回复一名用户或提供ID\\, 并提供权限\\, 以空格分隔\n支持的权限\\:\n%v", supportedPermissionsText),
			)
			return
		}
		var err error
		userIDStr = args[0]
		userID, err = strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			telegram.ReplyMessage(bot, message, "请不要输入奇怪的东西")
			return
		}
	} else {
		if message.ReplyToMessage.SenderChat != nil {
			userID = message.ReplyToMessage.SenderChat.ID
		} else {
			userID = message.ReplyToMessage.From.ID
		}
	}

	inputPermissions := make([]types.Permission, len(args)-1)
	unsupportedPermissions := make([]string, 0)
	for i, arg := range args[1:] {
		for _, p := range types.AllPermissions {
			if string(p) == arg {
				inputPermissions[i] = p
				break
			}
		}
		if inputPermissions[i] == "" {
			unsupportedPermissions = append(unsupportedPermissions, arg)
		}
	}

	if len(unsupportedPermissions) > 0 {
		telegram.ReplyMessageWithMarkdown(bot, message, common.EscapeMarkdown(fmt.Sprintf("权限不存在: %v\n支持的权限:\n", unsupportedPermissions))+supportedPermissionsText)
		return
	}

	isAdmin, err := service.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			isSuper := len(inputPermissions) == 0
			if strings.HasPrefix(userIDStr, "-100") && isSuper {
				telegram.ReplyMessage(bot, message, "禁止赋予群组所有权限")
				return
			}
			err := service.CreateOrUpdateAdmin(ctx, userID, inputPermissions, message.From.ID, isSuper)
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
		err := service.CreateOrUpdateAdmin(ctx, userID, inputPermissions, message.From.ID, false)
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
		originChannel, ok := telegram.GetMessageOriginChannelArtworkPost(ctx, bot, message)
		if !ok {
			telegram.ReplyMessage(bot, message, "用法:\n使用 /del 回复一条频道的图片消息, 或者提供频道消息ID.\n若使用 /delete 则删除整个作品")
			return
		}
		channelMessageID = originChannel.MessageID
	}
	if cmd == "del" {
		if !service.CheckAdminPermission(ctx, message.From.ID, types.PermissionDeleteArtwork) {
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
		telegram.ReplyMessage(bot, message, fmt.Sprintf("在数据库中已删除消息id为 %d 的图片", channelMessageID))
		bot.DeleteMessage(telegoutil.Delete(telegram.ChannelChatID, channelMessageID))

		if err := storage.GetStorage().DeletePicture(picture.StorageInfo); err != nil {
			Logger.Warnf("删除图片失败: %s", err)
			bot.SendMessage(telegoutil.Message(telegoutil.ID(message.From.ID), "在存储中删除图片文件失败: "+err.Error()))
		}
		return
	}
	if !service.CheckAdminPermission(ctx, message.From.ID, types.PermissionDeleteArtwork) {
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
	telegram.ReplyMessage(bot, message, fmt.Sprintf("在数据库中已删除消息id为 %d 的作品", channelMessageID))
	artworkMessageIDs := make([]int, len(artwork.Pictures))
	for i, picture := range artwork.Pictures {
		artworkMessageIDs[i] = picture.TelegramInfo.MessageID
	}
	bot.DeleteMessages(&telego.DeleteMessagesParams{
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
	if !service.CheckAdminPermission(ctx, message.From.ID, types.PermissionFetchArtwork) {
		telegram.ReplyMessage(bot, message, "你没有拉取作品的权限")
		return
	}

	go fetcher.FetchOnce(context.TODO(), config.Cfg.Fetcher.Limit)
	telegram.ReplyMessage(bot, message, "开始拉取作品了")
}

func postArtwork(ctx context.Context, bot *telego.Bot, query telego.CallbackQuery) {
	if !service.CheckAdminPermission(ctx, query.From.ID, types.PermissionPostArtwork) {
		bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "你没有发布作品的权限",
			ShowAlert:       true,
			CacheTime:       60,
		})
		return
	}
	queryDataSlice := strings.Split(query.Data, " ")
	asR18 := queryDataSlice[0] == "post_artwork_r18"
	dataID := queryDataSlice[1]
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

	var artwork *types.Artwork
	cachedArtwork, err := service.GetCachedArtworkByURL(ctx, sourceURL)
	if err != nil {
		artwork, err = sources.GetArtworkInfo(sourceURL)
		if err != nil {
			bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "获取作品信息失败" + err.Error(),
				ShowAlert:       true,
				CacheTime:       60,
			})
			return
		}
	} else {
		if cachedArtwork.Status == types.ArtworkStatusPosting {
			bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "该作品正在发布中",
				ShowAlert:       true,
				CacheTime:       60,
			})
			return
		}
		if err := service.UpdateCachedArtworkByURL(ctx, sourceURL, types.ArtworkStatusPosting); err != nil {
			Logger.Errorf("更新缓存作品状态失败: %s", err)
		}
		artwork = cachedArtwork.Artwork
		defer func() {
			if err := service.UpdateCachedArtworkByURL(ctx, sourceURL, types.ArtworkStatusCached); err != nil {
				Logger.Errorf("更新缓存作品状态失败: %s", err)
			}
		}()
	}
	go bot.EditMessageCaption(&telego.EditMessageCaptionParams{
		ChatID:      telegoutil.ID(query.Message.GetChat().ID),
		MessageID:   query.Message.GetMessageID(),
		Caption:     "正在发布...",
		ReplyMarkup: nil,
	})
	if service.CheckDeletedByURL(ctx, sourceURL) {
		if err := service.DeleteDeletedByURL(ctx, sourceURL); err != nil {
			Logger.Errorf("取消删除记录失败: %s", err)
			bot.EditMessageCaption(&telego.EditMessageCaptionParams{
				ChatID:    telegoutil.ID(query.Message.GetChat().ID),
				MessageID: query.Message.GetMessageID(),
				Caption:   "取消删除记录失败: " + err.Error(),
			})
			return
		}
	}
	if asR18 {
		artwork.R18 = true
	}
	if err := fetcher.PostAndCreateArtwork(ctx, artwork, bot, storage.GetStorage(), query.From.ID); err != nil {
		Logger.Errorf("发布失败: %s", err)
		bot.EditMessageCaption(&telego.EditMessageCaptionParams{
			ChatID:    telegoutil.ID(query.Message.GetChat().ID),
			MessageID: query.Message.GetMessageID(),
			Caption:   "发布失败: " + err.Error() + "\n\n" + time.Now().Format("2006-01-02 15:04:05"),
		})
		return
	}
	artwork, err = service.GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		Logger.Errorf("获取发布后的作品信息失败: %s", err)
		bot.EditMessageCaption(&telego.EditMessageCaptionParams{
			ChatID:      telegoutil.ID(query.Message.GetChat().ID),
			MessageID:   query.Message.GetMessageID(),
			Caption:     "发布成功, 但获取作品信息失败: " + err.Error(),
			ReplyMarkup: nil,
		})
		return
	}
	bot.EditMessageCaption(&telego.EditMessageCaptionParams{
		ChatID:    telegoutil.ID(query.Message.GetChat().ID),
		MessageID: query.Message.GetMessageID(),
		Caption:   "发布成功: " + artwork.Title + "\n\n发布时间: " + artwork.CreatedAt.Format("2006-01-02 15:04:05"),
		ReplyMarkup: telegoutil.InlineKeyboard(
			[]telego.InlineKeyboardButton{
				telegoutil.InlineKeyboardButton("查看").WithURL(telegram.GetArtworkPostMessageURL(artwork.Pictures[0].TelegramInfo.MessageID)),
				telegoutil.InlineKeyboardButton("原图").WithURL(telegram.GetDeepLinkForFile(artwork.Pictures[0].TelegramInfo.MessageID)),
			},
		),
	})
}

func processPictures(ctx context.Context, bot *telego.Bot, message telego.Message) {
	userAdmin, err := service.GetAdminByUserID(ctx, message.From.ID)
	if err != nil {
		telegram.ReplyMessage(bot, message, "获取管理员信息失败: "+err.Error())
		return
	}
	if userAdmin != nil && !userAdmin.SuperAdmin {
		telegram.ReplyMessage(bot, message, "你没有权限处理旧图片")
		return
	}
	go service.ProcessPicturesAndUpdate(context.TODO(), bot, &message)
	telegram.ReplyMessage(bot, message, "开始处理了")
}

func setArtworkR18(ctx context.Context, bot *telego.Bot, message telego.Message) {
	if !service.CheckAdminPermission(ctx, message.From.ID, types.PermissionEditArtwork) {
		telegram.ReplyMessage(bot, message, "你没有编辑作品的权限")
		return
	}

	messageOrigin, ok := telegram.GetMessageOriginChannelArtworkPost(ctx, bot, message)
	if !ok {
		telegram.ReplyMessage(bot, message, "请回复一条频道的图片消息")
		return
	}

	artwork, err := service.GetArtworkByMessageID(ctx, messageOrigin.MessageID)
	if err != nil {
		telegram.ReplyMessage(bot, message, "获取作品信息失败: "+err.Error())
		return
	}

	if err := service.UpdateArtworkR18ByURL(ctx, artwork.SourceURL, !artwork.R18); err != nil {
		telegram.ReplyMessage(bot, message, "更新作品信息失败: "+err.Error())
		return
	}
	telegram.ReplyMessage(bot, message, "该作品 R18 已标记为 "+strconv.FormatBool(!artwork.R18))
}

func setArtworkTags(ctx context.Context, bot *telego.Bot, message telego.Message) {
	if !service.CheckAdminPermission(ctx, message.From.ID, types.PermissionEditArtwork) {
		telegram.ReplyMessage(bot, message, "你没有编辑作品的权限")
		return
	}

	messageOrigin, ok := telegram.GetMessageOriginChannelArtworkPost(ctx, bot, message)
	if !ok {
		telegram.ReplyMessage(bot, message, "请回复一条频道的图片消息")
		return
	}

	artwork, err := service.GetArtworkByMessageID(ctx, messageOrigin.MessageID)
	if err != nil {
		telegram.ReplyMessage(bot, message, "获取作品信息失败: "+err.Error())
		return
	}

	cmd, _, args := telegoutil.ParseCommand(message.Text)
	if len(args) == 0 {
		telegram.ReplyMessage(bot, message, "请提供标签, 以空格分隔.\n不存在的标签将自动创建")
		return
	}
	tags := make([]string, 0)
	switch cmd {
	case "tags":
		tags = args
	case "addtags":
		tags = append(artwork.Tags, args...)
	case "deltags":
		tags = artwork.Tags[:]
		for _, arg := range args {
			for i, tag := range tags {
				if tag == arg {
					tags = append(tags[:i], tags[i+1:]...)
					break
				}
			}
		}
	}
	if err := service.UpdateArtworkTagsByURL(ctx, artwork.SourceURL, tags); err != nil {
		telegram.ReplyMessage(bot, message, "更新作品标签失败: "+err.Error())
		return
	}
	artwork, err = service.GetArtworkByURL(ctx, artwork.SourceURL)
	if err != nil {
		telegram.ReplyMessage(bot, message, "获取更新后的作品信息失败: "+err.Error())
		return
	}
	bot.EditMessageCaption(&telego.EditMessageCaptionParams{
		ChatID:    telegram.ChannelChatID,
		MessageID: artwork.Pictures[0].TelegramInfo.MessageID,
		Caption:   telegram.GetArtworkMarkdownCaption(artwork),
		ParseMode: telego.ModeMarkdownV2,
	})
	telegram.ReplyMessage(bot, message, "更新作品标签成功")
}
