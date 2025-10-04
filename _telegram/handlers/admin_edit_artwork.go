package handlers

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/internal/infra/database"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/service"
	"github.com/krau/ManyACG/sources"
	"github.com/krau/ManyACG/storage"
	"github.com/krau/ManyACG/telegram/utils"
	"github.com/krau/ManyACG/types"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func ToggleArtworkR18(ctx *telegohandler.Context, message telego.Message) error {
	if !CheckPermissionInGroup(ctx, message, shared.PermissionEditArtwork) {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "你没有编辑作品的权限")
		return nil
	}
	var sourceURL string
	if message.ReplyToMessage != nil {
		sourceURL = utils.FindSourceURLForMessage(message.ReplyToMessage)
	} else {
		sourceURL = sources.FindSourceURL(message.Text)
	}
	if sourceURL == "" {
		helpText := `
[管理员] <b>使用 /r18 命令回复一条包含作品链接的消息, 或在参数中提供作品链接, 将切换该作品的 R18 值</b>

命令语法: /r18 [作品链接]
`
		utils.ReplyMessageWithHTML(ctx, ctx.Bot(), message, helpText)
		return nil
	}

	artwork, err := database.Default().GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "获取作品信息失败: "+err.Error())
		return nil
	}
	if err := service.UpdateArtworkR18ByURL(ctx, sourceURL, !artwork.R18); err != nil {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "更新作品信息失败: "+err.Error())
		return nil
	}
	utils.ReplyMessage(ctx, ctx.Bot(), message, "该作品 R18 已标记为 "+strconv.FormatBool(!artwork.R18))
	return nil
}

func SetArtworkTags(ctx *telegohandler.Context, message telego.Message) error {
	if !CheckPermissionInGroup(ctx, message, shared.PermissionEditArtwork) {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "你没有编辑作品的权限")
		return nil
	}

	var sourceURL string
	if message.ReplyToMessage != nil {
		sourceURL = utils.FindSourceURLForMessage(message.ReplyToMessage)
	} else {
		sourceURL = sources.FindSourceURL(message.Text)
	}
	if sourceURL == "" {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "请回复一条消息, 或者指定作品链接")
		return nil
	}

	artwork, err := database.Default().GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "获取作品信息失败: "+err.Error())
		return nil
	}

	cmd, _, args := telegoutil.ParseCommand(message.Text)
	var argTags []string
	if message.ReplyToMessage != nil {
		if len(args) == 0 {
			utils.ReplyMessage(ctx, ctx.Bot(), message, "请提供标签, 用空格分隔")
			return nil
		}
		argTags = args
	} else {
		if len(args) <= 1 {
			utils.ReplyMessage(ctx, ctx.Bot(), message, "请在链接后提供标签, 用空格分隔")
			return nil
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
		utils.ReplyMessage(ctx, ctx.Bot(), message, "更新作品标签失败: "+err.Error())
		return nil
	}
	artwork, err = database.Default().GetArtworkByURL(ctx, artwork.SourceURL)
	if err != nil {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "获取更新后的作品信息失败: "+err.Error())
		return nil
	}
	if artwork.Pictures[0].TelegramInfo.MessageID != 0 {
		ctx.Bot().EditMessageCaption(ctx, &telego.EditMessageCaptionParams{
			ChatID:    ChannelChatID,
			MessageID: artwork.Pictures[0].TelegramInfo.MessageID,
			Caption:   utils.GetArtworkHTMLCaption(artwork),
			ParseMode: telego.ModeHTML,
		})
	}
	utils.ReplyMessage(ctx, ctx.Bot(), message, "更新作品标签成功")
	return nil
}

func EditArtworkR18(ctx *telegohandler.Context, query telego.CallbackQuery) error {
	if !CheckPermissionForQuery(ctx, query, shared.PermissionEditArtwork) {
		ctx.Bot().AnswerCallbackQuery(ctx,
			&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "你没有编辑作品的权限",
				ShowAlert:       true,
				CacheTime:       60,
			},
		)
		return nil
	}
	args := strings.Split(query.Data, " ")
	// edit_artwork r18 id 1
	if len(args) != 4 {
		ctx.Bot().AnswerCallbackQuery(ctx,
			&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "参数错误",
				ShowAlert:       true,
				CacheTime:       60,
			},
		)
		return nil
	}
	artworkID, err := primitive.ObjectIDFromHex(args[2])
	if err != nil {
		ctx.Bot().AnswerCallbackQuery(ctx,
			&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "无效的ID",
				ShowAlert:       true,
				CacheTime:       60,
			},
		)
		return nil
	}
	r18 := args[3] == "1"
	if err := service.UpdateArtworkR18ByID(ctx, artworkID, r18); err != nil {
		ctx.Bot().AnswerCallbackQuery(ctx,
			&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "更新作品信息失败: " + err.Error(),
				ShowAlert:       true,
				CacheTime:       60,
			},
		)
		return nil
	}
	ctx.Bot().AnswerCallbackQuery(ctx,
		&telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "该作品 R18 已标记为 " + strconv.FormatBool(r18),
			CacheTime:       2,
		},
	)
	return nil
}

func EditArtworkTitle(ctx *telegohandler.Context, message telego.Message) error {
	if !CheckPermissionInGroup(ctx, message, shared.PermissionEditArtwork) {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "你没有编辑作品的权限")
		return nil
	}

	var sourceURL string
	if message.ReplyToMessage != nil {
		sourceURL = utils.FindSourceURLForMessage(message.ReplyToMessage)
	} else {
		sourceURL = sources.FindSourceURL(message.Text)
	}
	if sourceURL == "" {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "请回复一条消息, 或者指定作品链接")
		return nil
	}

	artwork, err := database.Default().GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "获取作品信息失败: "+err.Error())
		return nil
	}

	_, _, args := telegoutil.ParseCommand(message.Text)
	var titleSlice []string
	if message.ReplyToMessage != nil {
		if len(args) == 0 {
			utils.ReplyMessage(ctx, ctx.Bot(), message, "请提供标题")
			return nil
		}
		titleSlice = args
	} else {
		if len(args) <= 1 {
			utils.ReplyMessage(ctx, ctx.Bot(), message, "请在链接后提供标题")
			return nil
		}
		titleSlice = args[1:]
	}
	title := strings.Join(titleSlice, " ")
	if err := service.UpdateArtworkTitleByURL(ctx, artwork.SourceURL, title); err != nil {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "更新作品标题失败: "+err.Error())
		return nil
	}
	artwork, err = database.Default().GetArtworkByURL(ctx, artwork.SourceURL)
	if err != nil {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "获取更新后的作品信息失败: "+err.Error())
		return nil
	}
	if artwork.Pictures[0].TelegramInfo.MessageID != 0 {
		ctx.Bot().EditMessageCaption(ctx, &telego.EditMessageCaptionParams{
			ChatID:    ChannelChatID,
			MessageID: artwork.Pictures[0].TelegramInfo.MessageID,
			Caption:   utils.GetArtworkHTMLCaption(artwork),
			ParseMode: telego.ModeHTML,
		})
	}
	utils.ReplyMessage(ctx, ctx.Bot(), message, "更新作品标题成功")
	return nil
}

// 删除 CachedArtwork, 刷新 telegram info
func RefreshArtwork(ctx *telegohandler.Context, message telego.Message) error {
	if !CheckPermissionInGroup(ctx, message, shared.PermissionEditArtwork) {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "你没有编辑作品的权限")
		return nil
	}

	var sourceURL string
	if message.ReplyToMessage != nil {
		sourceURL = utils.FindSourceURLForMessage(message.ReplyToMessage)
	} else {
		sourceURL = sources.FindSourceURL(message.Text)
	}
	if sourceURL == "" {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "请回复一条消息, 或者指定作品链接")
		return nil
	}

	if err := service.DeleteCachedArtworkByURL(ctx, sourceURL); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			common.Logger.Warnf("删除作品缓存失败: %s", err)
		} else {
			utils.ReplyMessage(ctx, ctx.Bot(), message, "删除作品缓存失败: "+err.Error())
			return nil
		}
	}

	artwork, err := database.Default().GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			utils.ReplyMessage(ctx, ctx.Bot(), message, "获取作品信息失败: "+err.Error())
			return nil
		}
		utils.ReplyMessage(ctx, ctx.Bot(), message, "该作品未发布, 缓存已删除")
		return nil
	}
	for _, picture := range artwork.Pictures {
		if picture.TelegramInfo == nil {
			continue
		}
		picture.TelegramInfo.PhotoFileID = ""
		picture.TelegramInfo.DocumentFileID = ""
		if err := service.UpdatePictureTelegramInfo(ctx, picture, picture.TelegramInfo); err != nil {
			utils.ReplyMessage(ctx, ctx.Bot(), message, "刷新作品信息失败: "+err.Error())
			return nil
		}
	}
	utils.ReplyMessage(ctx, ctx.Bot(), message, "已刷新作品信息")
	return nil
}

func ReCaptionArtwork(ctx *telegohandler.Context, message telego.Message) error {
	if !CheckPermissionInGroup(ctx, message, shared.PermissionEditArtwork) {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "你没有编辑作品的权限")
		return nil
	}
	var sourceURL string
	if message.ReplyToMessage != nil {
		sourceURL = utils.FindSourceURLForMessage(message.ReplyToMessage)
	} else {
		sourceURL = sources.FindSourceURL(message.Text)
	}
	if sourceURL == "" {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "请回复一条消息, 或者指定作品链接")
		return nil
	}
	artwork, err := database.Default().GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "获取作品信息失败: "+err.Error())
		return nil
	}
	if artwork.Pictures[0].TelegramInfo == nil || artwork.Pictures[0].TelegramInfo.MessageID == 0 {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "该作品未在频道发布")
		return nil
	}
	ctx.Bot().EditMessageCaption(ctx, &telego.EditMessageCaptionParams{
		ChatID:    ChannelChatID,
		MessageID: artwork.Pictures[0].TelegramInfo.MessageID,
		Caption:   utils.GetArtworkHTMLCaption(artwork),
		ParseMode: telego.ModeHTML,
	})
	utils.ReplyMessage(ctx, ctx.Bot(), message, "已重新生成作品描述")
	return nil
}

func AutoTaggingArtwork(ctx *telegohandler.Context, message telego.Message) error {
	if !CheckPermissionInGroup(ctx, message, shared.PermissionEditArtwork) {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "你没有编辑作品的权限")
		return nil
	}
	if common.TaggerClient == nil {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "Tagger is not available")
		return nil
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
		helpText := `
[管理员] <b>使用 /autotag 命令回复一条包含作品链接的消息, 或在参数中提供作品链接, 将基于AI自动为该作品添加标签</b>

命令语法: /autotag [作品链接] [图片序号]

若不提供参数, 默认选择所有图片
`
		utils.ReplyMessageWithHTML(ctx, ctx.Bot(), message, helpText)
		return nil
	}

	artwork, err := database.Default().GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "获取作品信息失败: "+err.Error())
		return nil
	}
	selectAllPictures := true
	pictureIndex := 1
	_, _, args := telegoutil.ParseCommand(message.Text)

	if len(args) > map[bool]int{true: 1, false: 0}[findUrlInArgs] {
		selectAllPictures = false
		pictureIndex, err := strconv.Atoi(args[len(args)-1])
		if err != nil {
			utils.ReplyMessage(ctx, ctx.Bot(), message, "图片序号错误")
			return nil
		}
		if pictureIndex < 1 || pictureIndex > len(artwork.Pictures) {
			utils.ReplyMessage(ctx, ctx.Bot(), message, "图片序号超出范围")
			return nil
		}
	}
	pictures := make([]*types.Picture, 0)
	if selectAllPictures {
		pictures = artwork.Pictures
	} else {
		picture := artwork.Pictures[pictureIndex-1]
		pictures[0] = picture
	}
	msg, err := utils.ReplyMessage(ctx, ctx.Bot(), message, "正在请求...")
	if err != nil {
		common.Logger.Errorf("Reply message failed: %s", err)
		return nil
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
			utils.ReplyMessage(ctx, ctx.Bot(), message, "Predict tags failed")
			return nil
		}
		if len(predict.PredictedTags) == 0 {
			utils.ReplyMessage(ctx, ctx.Bot(), message, "No tags predicted")
			return nil
		}
		newTags := slice.Union(artwork.Tags, predict.PredictedTags)
		if err := service.UpdateArtworkTagsByURL(ctx, artwork.SourceURL, newTags); err != nil {
			utils.ReplyMessage(ctx, ctx.Bot(), message, "更新作品标签失败: "+err.Error())
			return nil
		}
		artwork, err = database.Default().GetArtworkByURL(ctx, artwork.SourceURL)
		if err != nil {
			utils.ReplyMessage(ctx, ctx.Bot(), message, "获取更新后的作品信息失败: "+err.Error())
			return nil
		}
		if artwork.Pictures[0].TelegramInfo.MessageID != 0 {
			ctx.Bot().EditMessageCaption(ctx, &telego.EditMessageCaptionParams{
				ChatID:    ChannelChatID,
				MessageID: artwork.Pictures[0].TelegramInfo.MessageID,
				Caption:   utils.GetArtworkHTMLCaption(artwork),
				ParseMode: telego.ModeHTML,
			})
		}
		ctx.Bot().EditMessageText(ctx, &telego.EditMessageTextParams{
			ChatID:    msg.Chat.ChatID(),
			MessageID: msg.MessageID,
			Text:      fmt.Sprintf("选择的第 %d 张图片预测标签成功", i+1),
		})
	}
	ctx.Bot().EditMessageText(ctx, &telego.EditMessageTextParams{
		ChatID:    msg.Chat.ChatID(),
		MessageID: msg.MessageID,
		Text:      "更新作品标签成功",
	})
	return nil
}
