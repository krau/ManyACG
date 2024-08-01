package handlers

import (
	. "ManyACG/logger"
	"ManyACG/service"
	"ManyACG/storage"
	"ManyACG/telegram"
	"ManyACG/types"
	"context"
	"fmt"
	"strconv"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

func DeletePicture(ctx context.Context, bot *telego.Bot, message telego.Message) {
	if !CheckPermissionInGroup(ctx, message, types.PermissionDeleteArtwork) {
		telegram.ReplyMessage(bot, message, "你没有删除图片的权限")
		return
	}
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
