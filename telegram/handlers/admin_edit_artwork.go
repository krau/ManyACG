package handlers

import (
	"ManyACG/service"
	"ManyACG/telegram/utils"
	"ManyACG/types"
	"context"
	"strconv"
	"strings"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func SetArtworkR18(ctx context.Context, bot *telego.Bot, message telego.Message) {
	if !CheckPermissionInGroup(ctx, message, types.PermissionEditArtwork) {
		utils.ReplyMessage(bot, message, "你没有编辑作品的权限")
		return
	}
	messageOrigin, ok := utils.GetMessageOriginChannelArtworkPost(ctx, bot, message)
	if !ok {
		utils.ReplyMessage(bot, message, "请回复一条频道的图片消息")
		return
	}

	artwork, err := service.GetArtworkByMessageID(ctx, messageOrigin.MessageID)
	if err != nil {
		utils.ReplyMessage(bot, message, "获取作品信息失败: "+err.Error())
		return
	}

	if err := service.UpdateArtworkR18ByURL(ctx, artwork.SourceURL, !artwork.R18); err != nil {
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
	messageOrigin, ok := utils.GetMessageOriginChannelArtworkPost(ctx, bot, message)
	if !ok {
		utils.ReplyMessage(bot, message, "请回复一条频道的图片消息")
		return
	}

	artwork, err := service.GetArtworkByMessageID(ctx, messageOrigin.MessageID)
	if err != nil {
		utils.ReplyMessage(bot, message, "获取作品信息失败: "+err.Error())
		return
	}

	cmd, _, args := telegoutil.ParseCommand(message.Text)
	if len(args) == 0 {
		utils.ReplyMessage(bot, message, "请提供标签, 以空格分隔.\n不存在的标签将自动创建")
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
	for i, tag := range tags {
		tags[i] = strings.TrimPrefix(tag, "#")
	}
	if err := service.UpdateArtworkTagsByURL(ctx, artwork.SourceURL, tags); err != nil {
		utils.ReplyMessage(bot, message, "更新作品标签失败: "+err.Error())
		return
	}
	artwork, err = service.GetArtworkByURL(ctx, artwork.SourceURL)
	if err != nil {
		utils.ReplyMessage(bot, message, "获取更新后的作品信息失败: "+err.Error())
		return
	}
	bot.EditMessageCaption(&telego.EditMessageCaptionParams{
		ChatID:    ChannelChatID,
		MessageID: artwork.Pictures[0].TelegramInfo.MessageID,
		Caption:   utils.GetArtworkHTMLCaption(artwork),
		ParseMode: telego.ModeHTML,
	})
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
