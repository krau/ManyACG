package handlers

import (
	"context"
	"strings"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/service"
	"github.com/krau/ManyACG/telegram/utils"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Start(ctx context.Context, bot *telego.Bot, message telego.Message) {
	_, _, args := telegoutil.ParseCommand(message.Text)
	if len(args) > 0 {
		common.Logger.Debugf("start: args=%v", args)
		action := strings.Split(args[0], "_")[0]
		switch action {
		case "file":
			pictureID := args[0][5:]
			_, err := utils.SendPictureFileByID(ctx, bot, message, ChannelChatID, pictureID)
			if err != nil {
				utils.ReplyMessage(bot, message, "获取失败: "+err.Error())
			}
		case "files":
			artworkID := args[0][6:]
			objectID, err := primitive.ObjectIDFromHex(artworkID)
			if err != nil {
				utils.ReplyMessage(bot, message, "无效的ID")
				return
			}
			artwork, err := service.GetArtworkByID(ctx, objectID)
			if err != nil {
				utils.ReplyMessage(bot, message, "获取失败: "+err.Error())
				return
			}
			getArtworkFiles(ctx, bot, message, artwork)
		case "code":
			userID := message.From.ID
			userModel, _ := service.GetUserByTelegramID(ctx, userID)
			if userModel != nil {
				bot.SendMessage(telegoutil.Messagef(message.Chat.ChatID(), "您的此 Telegram 账号 ( %d ) 已经绑定了 ManyACG 账号 %s", userID, userModel.Username))
				return
			}
			unauthUserID := args[0][5:]
			objectID, err := primitive.ObjectIDFromHex(unauthUserID)
			if err != nil {
				utils.ReplyMessage(bot, message, "无效的ID")
				return
			}
			unauthUser, err := service.GetUnauthUserByID(ctx, objectID)
			if err != nil {
				utils.ReplyMessage(bot, message, "获取失败: "+err.Error())
				return
			}
			_, err = bot.SendMessage(telegoutil.Messagef(message.Chat.ChatID(),
				"您的此 Telegram 账号 ( %d ) 将与 ManyACG 账号 %s 绑定\n验证码: <code>%s</code>",
				userID,
				common.EscapeHTML(unauthUser.Username),
				common.EscapeHTML(unauthUser.Code)).
				WithParseMode(telego.ModeHTML),
			)
			if err != nil {
				common.Logger.Errorf("Failed to send message: %v", err)
				return
			}
			unauthUser.TelegramID = userID
			service.UpdateUnauthUser(ctx, objectID, unauthUser)
		case "info":
			dataID := args[0][5:]
			sourceURL, err := service.GetCallbackDataByID(ctx, dataID)
			if err != nil {
				utils.ReplyMessage(bot, message, "获取失败: "+err.Error())
				return
			}
			if err := utils.SendFullArtworkInfo(ctx, bot, message, sourceURL); err != nil {
				utils.ReplyMessage(bot, message, err.Error())
			}
		}
		return
	}
	Help(ctx, bot, message)
}
