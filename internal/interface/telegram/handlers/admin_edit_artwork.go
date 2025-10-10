package handlers

import (
	"strconv"
	"strings"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/krau/ManyACG/internal/interface/telegram/handlers/utils"
	"github.com/krau/ManyACG/internal/interface/telegram/metautil"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/samber/oops"
)

func ToggleArtworkR18(ctx *telegohandler.Context, message telego.Message) error {
	serv := service.FromContext(ctx)
	if !utils.CheckPermissionInGroup(ctx, serv, message, shared.PermissionEditArtwork) {
		utils.ReplyMessage(ctx, message, "你没有编辑作品的权限")
		return nil
	}
	var sourceURL string
	if message.ReplyToMessage != nil {
		sourceURL = utils.FindSourceURLInMessage(serv, message.ReplyToMessage)
	} else {
		sourceURL = serv.FindSourceURL(message.Text)
	}
	if sourceURL == "" {
		helpText := `
[管理员] <b>使用 /r18 命令回复一条包含作品链接的消息, 或在参数中提供作品链接, 将切换该作品的 R18 值</b>

命令语法: /r18 [作品链接]
`
		utils.ReplyMessageWithHTML(ctx, message, helpText)
		return nil
	}

	artwork, err := serv.GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		utils.ReplyMessage(ctx, message, "获取作品信息失败: "+err.Error())
		return nil
	}
	if err := serv.UpdateArtworkR18ByURL(ctx, sourceURL, !artwork.R18); err != nil {
		utils.ReplyMessage(ctx, message, "更新作品信息失败: "+err.Error())
		return nil
	}
	utils.ReplyMessage(ctx, message, "该作品 R18 已标记为 "+strconv.FormatBool(!artwork.R18))
	return nil
}

func SetArtworkTags(ctx *telegohandler.Context, message telego.Message) error {
	serv := service.FromContext(ctx)
	if !utils.CheckPermissionInGroup(ctx, serv, message, shared.PermissionEditArtwork) {
		utils.ReplyMessage(ctx, message, "你没有编辑作品的权限")
		return nil
	}

	var sourceURL string
	if message.ReplyToMessage != nil {
		sourceURL = utils.FindSourceURLInMessage(serv, message.ReplyToMessage)
	} else {
		sourceURL = serv.FindSourceURL(message.Text)
	}
	if sourceURL == "" {
		utils.ReplyMessage(ctx, message, "请回复一条消息, 或者指定作品链接")
		return nil
	}

	artwork, err := serv.GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		utils.ReplyMessage(ctx, message, "获取作品信息失败: "+err.Error())
		return nil
	}

	cmd, _, args := telegoutil.ParseCommand(message.Text)
	var argTags []string
	if message.ReplyToMessage != nil {
		if len(args) == 0 {
			utils.ReplyMessage(ctx, message, "请提供标签, 用空格分隔")
			return nil
		}
		argTags = args
	} else {
		if len(args) <= 1 {
			utils.ReplyMessage(ctx, message, "请在链接后提供标签, 用空格分隔")
			return nil
		}
		argTags = args[1:]
	}

	newTags := make([]string, 0)
	origTags := make([]string, len(artwork.Tags))
	for i, tag := range artwork.Tags {
		origTags[i] = tag.Name
	}
	switch cmd {
	case "tags":
		newTags = argTags
	case "addtags":
		newTags = append(origTags, argTags...)
	case "deltags":
		newTags = origTags[:]
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
	newTags = slice.Unique(newTags)

	if err := serv.UpdateArtworkTagsByURL(ctx, artwork.SourceURL, newTags); err != nil {
		utils.ReplyMessage(ctx, message, "更新作品标签失败: "+err.Error())
		return nil
	}
	artwork, err = serv.GetArtworkByURL(ctx, artwork.SourceURL)
	if err != nil {
		utils.ReplyMessage(ctx, message, "获取更新后的作品信息失败: "+err.Error())
		return nil
	}
	if msgId := artwork.Pictures[0].TelegramInfo.Data().MessageID; msgId != 0 {
		channelMeta := metautil.FromContext(ctx)
		ctx.Bot().EditMessageCaption(ctx, &telego.EditMessageCaptionParams{
			ChatID:    channelMeta.ChannelChatID(),
			MessageID: msgId,
			Caption:   utils.ArtworkHTMLCaption(artwork),
			ParseMode: telego.ModeHTML,
		})
	}
	utils.ReplyMessage(ctx, message, "更新作品标签成功")
	return nil
}

func EditArtworkR18(ctx *telegohandler.Context, query telego.CallbackQuery) error {
	serv := service.FromContext(ctx)
	if !utils.CheckPermissionForQuery(ctx, serv, query, shared.PermissionEditArtwork) {
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
	artworkID, err := objectuuid.FromObjectIDHex(args[2])
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
	if err := serv.UpdateArtworkR18ByID(ctx, artworkID, r18); err != nil {
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
	serv := service.FromContext(ctx)
	if !utils.CheckPermissionInGroup(ctx, serv, message, shared.PermissionEditArtwork) {
		utils.ReplyMessage(ctx, message, "你没有编辑作品的权限")
		return nil
	}

	var sourceURL string
	if message.ReplyToMessage != nil {
		sourceURL = utils.FindSourceURLInMessage(serv, message.ReplyToMessage)
	} else {
		sourceURL = serv.FindSourceURL(message.Text)
	}
	if sourceURL == "" {
		utils.ReplyMessage(ctx, message, "请回复一条消息, 或者指定作品链接")
		return nil
	}

	artwork, err := serv.GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		utils.ReplyMessage(ctx, message, "获取作品信息失败: "+err.Error())
		return nil
	}

	_, _, args := telegoutil.ParseCommand(message.Text)
	var titleSlice []string
	if message.ReplyToMessage != nil {
		if len(args) == 0 {
			utils.ReplyMessage(ctx, message, "请提供标题")
			return nil
		}
		titleSlice = args
	} else {
		if len(args) <= 1 {
			utils.ReplyMessage(ctx, message, "请在链接后提供标题")
			return nil
		}
		titleSlice = args[1:]
	}
	title := strings.Join(titleSlice, " ")
	if err := serv.UpdateArtworkTitleByURL(ctx, artwork.SourceURL, title); err != nil {
		utils.ReplyMessage(ctx, message, "更新作品标题失败: "+err.Error())
		return nil
	}
	artwork, err = serv.GetArtworkByURL(ctx, artwork.SourceURL)
	if err != nil {
		utils.ReplyMessage(ctx, message, "获取更新后的作品信息失败: "+err.Error())
		return nil
	}
	if msgId := artwork.Pictures[0].TelegramInfo.Data().MessageID; msgId != 0 {
		meta := metautil.FromContext(ctx)
		ctx.Bot().EditMessageCaption(ctx, &telego.EditMessageCaptionParams{
			ChatID:    meta.ChannelChatID(),
			MessageID: msgId,
			Caption:   utils.ArtworkHTMLCaption(artwork),
			ParseMode: telego.ModeHTML,
		})
	}
	utils.ReplyMessage(ctx, message, "更新作品标题成功")
	return nil
}

// 删除 CachedArtwork, 刷新 telegram info
func RefreshArtwork(ctx *telegohandler.Context, message telego.Message) error {
	serv := service.FromContext(ctx)
	if !utils.CheckPermissionInGroup(ctx, serv, message, shared.PermissionEditArtwork) {
		utils.ReplyMessage(ctx, message, "你没有编辑作品的权限")
		return nil
	}

	var sourceURL string
	if message.ReplyToMessage != nil {
		sourceURL = utils.FindSourceURLInMessage(serv, message.ReplyToMessage)
	} else {
		sourceURL = serv.FindSourceURL(message.Text)
	}
	if sourceURL == "" {
		utils.ReplyMessage(ctx, message, "请回复一条消息, 或者指定作品链接")
		return nil
	}

	if err := serv.DeleteCachedArtworkByURL(ctx, sourceURL); err != nil {
		utils.ReplyMessage(ctx, message, "删除作品缓存失败: "+err.Error())
		return nil
	}

	artwork, err := serv.GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		utils.ReplyMessage(ctx, message, "获取作品信息失败: "+err.Error())
		return nil
	}
	for _, picture := range artwork.Pictures {
		newInfo := &shared.TelegramInfo{
			MessageID:    picture.TelegramInfo.Data().MessageID,
			MediaGroupID: picture.TelegramInfo.Data().MediaGroupID,
		}
		if err := serv.UpdatePictureTelegramInfo(ctx, picture.ID, newInfo); err != nil {
			utils.ReplyMessage(ctx, message, "刷新作品信息失败: "+err.Error())
			return nil
		}
	}
	utils.ReplyMessage(ctx, message, "已刷新作品信息")
	return nil
}

func ReCaptionArtwork(ctx *telegohandler.Context, message telego.Message) error {
	serv := service.FromContext(ctx)
	if !utils.CheckPermissionInGroup(ctx, serv, message, shared.PermissionEditArtwork) {
		utils.ReplyMessage(ctx, message, "你没有编辑作品的权限")
		return nil
	}
	var sourceURL string
	if message.ReplyToMessage != nil {
		sourceURL = utils.FindSourceURLInMessage(serv, message.ReplyToMessage)
	} else {
		sourceURL = serv.FindSourceURL(message.Text)
	}
	if sourceURL == "" {
		utils.ReplyMessage(ctx, message, "请回复一条消息, 或者指定作品链接")
		return nil
	}
	artwork, err := serv.GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		utils.ReplyMessage(ctx, message, "获取作品信息失败: "+err.Error())
		return nil
	}
	if artwork.Pictures[0].TelegramInfo.Data().MessageID == 0 {
		utils.ReplyMessage(ctx, message, "该作品未在频道发布")
		return nil
	}
	meta := metautil.FromContext(ctx)
	ctx.Bot().EditMessageCaption(ctx, &telego.EditMessageCaptionParams{
		ChatID:    meta.ChannelChatID(),
		MessageID: artwork.Pictures[0].TelegramInfo.Data().MessageID,
		Caption:   utils.ArtworkHTMLCaption(artwork),
		ParseMode: telego.ModeHTML,
	})
	utils.ReplyMessage(ctx, message, "已重新生成作品描述")
	return nil
}

func AutoTaggingArtwork(ctx *telegohandler.Context, message telego.Message) error {
	serv := service.FromContext(ctx)
	if !utils.CheckPermissionInGroup(ctx, serv, message, shared.PermissionEditArtwork) {
		utils.ReplyMessage(ctx, message, "你没有编辑作品的权限")
		return nil
	}
	var sourceURL string
	if message.ReplyToMessage != nil {
		sourceURL = utils.FindSourceURLInMessage(serv, message.ReplyToMessage)
	} else {
		sourceURL = serv.FindSourceURL(message.Text)
	}
	if sourceURL == "" {
		helpText := `
[管理员] <b>使用 /autotag 命令回复一条包含作品链接的消息, 或在参数中提供作品链接, 将基于AI自动为该作品添加标签</b>

命令语法: /autotag [作品链接]
`
		utils.ReplyMessageWithHTML(ctx, message, helpText)
		return nil
	}

	artwork, err := serv.GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		utils.ReplyMessage(ctx, message, "获取作品信息失败: "+err.Error())
		return nil
	}
	msg, err := utils.ReplyMessage(ctx, message, "正在请求...")
	if err != nil {
		return oops.Wrapf(err, "failed to send message")
	}
	if err := serv.PredictAndUpdateArtworkTags(ctx, artwork.ID); err != nil {
		log.Errorf("failed to predict and update artwork tags: %s", err)
		ctx.Bot().EditMessageText(ctx, &telego.EditMessageTextParams{
			ChatID:    msg.Chat.ChatID(),
			MessageID: msg.MessageID,
			Text:      "自动添加作品标签失败: " + err.Error(),
		})
		return nil
	}
	newAw, err := serv.GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		log.Errorf("failed to get artwork by url after autotag: %s", err)
		ctx.Bot().EditMessageText(ctx, &telego.EditMessageTextParams{
			ChatID:    msg.Chat.ChatID(),
			MessageID: msg.MessageID,
			Text:      "获取更新后的作品信息失败: " + err.Error(),
		})
		return nil
	}
	if msgId := newAw.Pictures[0].TelegramInfo.Data().MessageID; msgId != 0 {
		meta := metautil.FromContext(ctx)
		caption := utils.ArtworkHTMLCaption(newAw)
		ctx.Bot().EditMessageCaption(ctx, &telego.EditMessageCaptionParams{
			ChatID:    meta.ChannelChatID(),
			MessageID: msgId,
			Caption:   caption,
			ParseMode: telego.ModeHTML,
		})
	}
	ctx.Bot().EditMessageText(ctx, &telego.EditMessageTextParams{
		ChatID:    msg.Chat.ChatID(),
		MessageID: msg.MessageID,
		Text:      "更新作品标签成功",
	})
	return nil
}
