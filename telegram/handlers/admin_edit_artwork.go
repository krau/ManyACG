package handlers

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/service"
	"github.com/krau/ManyACG/sources"
	"github.com/krau/ManyACG/storage"
	"github.com/krau/ManyACG/telegram/utils"
	"github.com/krau/ManyACG/types"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func ToggleArtworkR18(ctx context.Context, bot *telego.Bot, message telego.Message) {
	if !CheckPermissionInGroup(ctx, message, types.PermissionEditArtwork) {
		utils.ReplyMessage(bot, message, "你没有编辑作品的权限")
		return
	}
	var sourceURL string
	if message.ReplyToMessage != nil {
		sourceURL = utils.FindSourceURLForMessage(message.ReplyToMessage)
	} else {
		sourceURL = sources.FindSourceURL(message.Text)
	}
	if sourceURL == "" {
		utils.ReplyMessage(bot, message, "请回复一条消息, 或者指定作品链接")
		return
	}

	artwork, err := service.GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		utils.ReplyMessage(bot, message, "获取作品信息失败: "+err.Error())
		return
	}
	if err := service.UpdateArtworkR18ByURL(ctx, sourceURL, !artwork.R18); err != nil {
		utils.ReplyMessage(bot, message, "更新作品信息失败: "+err.Error())
		return
	}
	utils.ReplyMessage(bot, message, "该作品 R18 已标记为 "+strconv.FormatBool(!artwork.R18))
}

func SetArtworkTags(ctx context.Context, bot *telego.Bot, message telego.Message) {
	if !CheckPermissionInGroup(ctx, message, types.PermissionEditArtwork) {
		utils.ReplyMessage(bot, message, "你没有编辑作品的权限")
		return
	}

	var sourceURL string
	if message.ReplyToMessage != nil {
		sourceURL = utils.FindSourceURLForMessage(message.ReplyToMessage)
	} else {
		sourceURL = sources.FindSourceURL(message.Text)
	}
	if sourceURL == "" {
		utils.ReplyMessage(bot, message, "请回复一条消息, 或者指定作品链接")
		return
	}

	artwork, err := service.GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		utils.ReplyMessage(bot, message, "获取作品信息失败: "+err.Error())
		return
	}

	cmd, _, args := telegoutil.ParseCommand(message.Text)
	var argTags []string
	if message.ReplyToMessage != nil {
		if len(args) == 0 {
			utils.ReplyMessage(bot, message, "请提供标签, 用空格分隔")
			return
		}
		argTags = args
	} else {
		if len(args) <= 1 {
			utils.ReplyMessage(bot, message, "请在链接后提供标签, 用空格分隔")
			return
		}
		argTags = args[1:]
	}

	newTags := make([]string, 0)
	switch cmd {
	case "tags":
		newTags = argTags
	case "addtags":
		newTags = append(artwork.Tags, argTags...)
	case "deltags":
		newTags = artwork.Tags[:]
		for _, arg := range argTags {
			for i, tag := range newTags {
				if tag == arg {
					newTags = append(newTags[:i], newTags[i+1:]...)
					break
				}
			}
		}
	}
	for i, tag := range newTags {
		newTags[i] = strings.TrimPrefix(tag, "#")
	}
	if err := service.UpdateArtworkTagsByURL(ctx, artwork.SourceURL, newTags); err != nil {
		utils.ReplyMessage(bot, message, "更新作品标签失败: "+err.Error())
		return
	}
	artwork, err = service.GetArtworkByURL(ctx, artwork.SourceURL)
	if err != nil {
		utils.ReplyMessage(bot, message, "获取更新后的作品信息失败: "+err.Error())
		return
	}
	if artwork.Pictures[0].TelegramInfo.MessageID != 0 {
		bot.EditMessageCaption(&telego.EditMessageCaptionParams{
			ChatID:    ChannelChatID,
			MessageID: artwork.Pictures[0].TelegramInfo.MessageID,
			Caption:   utils.GetArtworkHTMLCaption(artwork),
			ParseMode: telego.ModeHTML,
		})
	}
	utils.ReplyMessage(bot, message, "更新作品标签成功")
}

func EditArtworkR18(ctx context.Context, bot *telego.Bot, query telego.CallbackQuery) {
	if !CheckPermissionForQuery(ctx, query, types.PermissionEditArtwork) {
		bot.AnswerCallbackQuery(
			&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "你没有编辑作品的权限",
				ShowAlert:       true,
				CacheTime:       60,
			},
		)
		return
	}
	args := strings.Split(query.Data, " ")
	// edit_artwork r18 id 1
	if len(args) != 4 {
		bot.AnswerCallbackQuery(
			&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "参数错误",
				ShowAlert:       true,
				CacheTime:       60,
			},
		)
		return
	}
	artworkID, err := primitive.ObjectIDFromHex(args[2])
	if err != nil {
		bot.AnswerCallbackQuery(
			&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "无效的ID",
				ShowAlert:       true,
				CacheTime:       60,
			},
		)
		return
	}
	r18 := args[3] == "1"
	if err := service.UpdateArtworkR18ByID(ctx, artworkID, r18); err != nil {
		bot.AnswerCallbackQuery(
			&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "更新作品信息失败: " + err.Error(),
				ShowAlert:       true,
				CacheTime:       60,
			},
		)
		return
	}
	bot.AnswerCallbackQuery(
		&telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "该作品 R18 已标记为 " + strconv.FormatBool(r18),
			CacheTime:       2,
		},
	)
}

func EditArtworkTitle(ctx context.Context, bot *telego.Bot, message telego.Message) {
	if !CheckPermissionInGroup(ctx, message, types.PermissionEditArtwork) {
		utils.ReplyMessage(bot, message, "你没有编辑作品的权限")
		return
	}

	var sourceURL string
	if message.ReplyToMessage != nil {
		sourceURL = utils.FindSourceURLForMessage(message.ReplyToMessage)
	} else {
		sourceURL = sources.FindSourceURL(message.Text)
	}
	if sourceURL == "" {
		utils.ReplyMessage(bot, message, "请回复一条消息, 或者指定作品链接")
		return
	}

	artwork, err := service.GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		utils.ReplyMessage(bot, message, "获取作品信息失败: "+err.Error())
		return
	}

	_, _, args := telegoutil.ParseCommand(message.Text)
	var titleSlice []string
	if message.ReplyToMessage != nil {
		if len(args) == 0 {
			utils.ReplyMessage(bot, message, "请提供标题")
			return
		}
		titleSlice = args
	} else {
		if len(args) <= 1 {
			utils.ReplyMessage(bot, message, "请在链接后提供标题")
			return
		}
		titleSlice = args[1:]
	}
	title := strings.Join(titleSlice, " ")
	if err := service.UpdateArtworkTitleByURL(ctx, artwork.SourceURL, title); err != nil {
		utils.ReplyMessage(bot, message, "更新作品标题失败: "+err.Error())
		return
	}
	artwork, err = service.GetArtworkByURL(ctx, artwork.SourceURL)
	if err != nil {
		utils.ReplyMessage(bot, message, "获取更新后的作品信息失败: "+err.Error())
		return
	}
	if artwork.Pictures[0].TelegramInfo.MessageID != 0 {
		bot.EditMessageCaption(&telego.EditMessageCaptionParams{
			ChatID:    ChannelChatID,
			MessageID: artwork.Pictures[0].TelegramInfo.MessageID,
			Caption:   utils.GetArtworkHTMLCaption(artwork),
			ParseMode: telego.ModeHTML,
		})
	}
	utils.ReplyMessage(bot, message, "更新作品标题成功")
}

// 删除 CachedArtwork, 刷新 telegram info
func RefreshArtwork(ctx context.Context, bot *telego.Bot, message telego.Message) {
	if !CheckPermissionInGroup(ctx, message, types.PermissionEditArtwork) {
		utils.ReplyMessage(bot, message, "你没有编辑作品的权限")
		return
	}

	var sourceURL string
	if message.ReplyToMessage != nil {
		sourceURL = utils.FindSourceURLForMessage(message.ReplyToMessage)
	} else {
		sourceURL = sources.FindSourceURL(message.Text)
	}
	if sourceURL == "" {
		utils.ReplyMessage(bot, message, "请回复一条消息, 或者指定作品链接")
		return
	}

	if err := service.DeleteCachedArtworkByURL(ctx, sourceURL); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			common.Logger.Warnf("删除作品缓存失败: %s", err)
		} else {
			utils.ReplyMessage(bot, message, "删除作品缓存失败: "+err.Error())
			return
		}
	}

	artwork, err := service.GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			utils.ReplyMessage(bot, message, "获取作品信息失败: "+err.Error())
			return
		}
		utils.ReplyMessage(bot, message, "该作品未发布, 缓存已删除")
		return
	}
	for _, picture := range artwork.Pictures {
		if picture.TelegramInfo == nil {
			continue
		}
		picture.TelegramInfo.PhotoFileID = ""
		picture.TelegramInfo.DocumentFileID = ""
		if err := service.UpdatePictureTelegramInfo(ctx, picture, picture.TelegramInfo); err != nil {
			utils.ReplyMessage(bot, message, "刷新作品信息失败: "+err.Error())
			return
		}
	}
	utils.ReplyMessage(bot, message, "已刷新作品信息")
}

func ReCaptionArtwork(ctx context.Context, bot *telego.Bot, message telego.Message) {
	if !CheckPermissionInGroup(ctx, message, types.PermissionEditArtwork) {
		utils.ReplyMessage(bot, message, "你没有编辑作品的权限")
		return
	}
	var sourceURL string
	if message.ReplyToMessage != nil {
		sourceURL = utils.FindSourceURLForMessage(message.ReplyToMessage)
	} else {
		sourceURL = sources.FindSourceURL(message.Text)
	}
	if sourceURL == "" {
		utils.ReplyMessage(bot, message, "请回复一条消息, 或者指定作品链接")
		return
	}
	artwork, err := service.GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		utils.ReplyMessage(bot, message, "获取作品信息失败: "+err.Error())
		return
	}
	if artwork.Pictures[0].TelegramInfo == nil || artwork.Pictures[0].TelegramInfo.MessageID == 0 {
		utils.ReplyMessage(bot, message, "该作品未在频道发布")
		return
	}
	bot.EditMessageCaption(&telego.EditMessageCaptionParams{
		ChatID:    ChannelChatID,
		MessageID: artwork.Pictures[0].TelegramInfo.MessageID,
		Caption:   utils.GetArtworkHTMLCaption(artwork),
		ParseMode: telego.ModeHTML,
	})
	utils.ReplyMessage(bot, message, "已重新生成作品描述")
}

func AutoTaggingArtwork(ctx context.Context, bot *telego.Bot, message telego.Message) {
	if !CheckPermissionInGroup(ctx, message, types.PermissionEditArtwork) {
		utils.ReplyMessage(bot, message, "你没有编辑作品的权限")
		return
	}
	if common.TaggerClient == nil {
		utils.ReplyMessage(bot, message, "Tagger is not available")
		return
	}
	var sourceURL string
	var findUrlInArgs bool
	if message.ReplyToMessage != nil {
		sourceURL = utils.FindSourceURLForMessage(message.ReplyToMessage)
	} else {
		sourceURL = sources.FindSourceURL(message.Text)
		findUrlInArgs = true
	}
	if sourceURL == "" {
		utils.ReplyMessage(bot, message, "请回复一条消息, 或者指定作品链接")
		return
	}

	artwork, err := service.GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		utils.ReplyMessage(bot, message, "获取作品信息失败: "+err.Error())
		return
	}
	selectAllPictures := true
	pictureIndex := 1
	_, _, args := telegoutil.ParseCommand(message.Text)
	if len(args) > func() int {
		if findUrlInArgs {
			return 1
		}
		return 0
	}() {
		selectAllPictures = false
		pictureIndex, err := strconv.Atoi(args[len(args)-1])
		if err != nil {
			utils.ReplyMessage(bot, message, "图片序号错误")
			return
		}
		if pictureIndex < 1 || pictureIndex > len(artwork.Pictures) {
			utils.ReplyMessage(bot, message, "图片序号超出范围")
			return
		}
	}
	pictures := make([]*types.Picture, 0)
	if selectAllPictures {
		pictures = artwork.Pictures
	} else {
		picture := artwork.Pictures[pictureIndex-1]
		pictures[0] = picture
	}
	msg, err := utils.ReplyMessage(bot, message, "正在请求...")
	if err != nil {
		common.Logger.Errorf("Reply message failed: %s", err)
		return
	}
	for i, picture := range pictures {
		var file []byte
		file, err = storage.GetFile(ctx, func() *types.StorageDetail {
			if picture.StorageInfo.Regular != nil {
				return picture.StorageInfo.Regular
			} else {
				return picture.StorageInfo.Original
			}
		}())
		if err != nil {
			common.Logger.Errorf("Get picture file from storage failed: %s", err)
			file, err = common.DownloadWithCache(ctx, picture.Original, nil)
		}
		if err != nil {
			common.Logger.Errorf("Download picture %s failed: %s", picture.Original, err)
			continue
		}
		common.Logger.Debugf("Predicting tags for %s", picture.Original)
		predict, err := common.TaggerClient.Predict(ctx, file)
		if err != nil {
			common.Logger.Errorf("Predict tags failed: %s", err)
			utils.ReplyMessage(bot, message, "Predict tags failed")
			return
		}
		if len(predict.PredictedTags) == 0 {
			utils.ReplyMessage(bot, message, "No tags predicted")
			return
		}
		newTags := slice.Union(artwork.Tags, predict.PredictedTags)
		if err := service.UpdateArtworkTagsByURL(ctx, artwork.SourceURL, newTags); err != nil {
			utils.ReplyMessage(bot, message, "更新作品标签失败: "+err.Error())
			return
		}
		artwork, err = service.GetArtworkByURL(ctx, artwork.SourceURL)
		if err != nil {
			utils.ReplyMessage(bot, message, "获取更新后的作品信息失败: "+err.Error())
			return
		}
		if artwork.Pictures[0].TelegramInfo.MessageID != 0 {
			bot.EditMessageCaption(&telego.EditMessageCaptionParams{
				ChatID:    ChannelChatID,
				MessageID: artwork.Pictures[0].TelegramInfo.MessageID,
				Caption:   utils.GetArtworkHTMLCaption(artwork),
				ParseMode: telego.ModeHTML,
			})
		}
		bot.EditMessageText(&telego.EditMessageTextParams{
			ChatID:    msg.Chat.ChatID(),
			MessageID: msg.MessageID,
			Text:      fmt.Sprintf("选择的第 %d 张图片预测标签成功", i+1),
		})
	}
	bot.EditMessageText(&telego.EditMessageTextParams{
		ChatID:    msg.Chat.ChatID(),
		MessageID: msg.MessageID,
		Text:      "更新作品标签成功",
	})
}
