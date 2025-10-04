package handlers

import (
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
)

func Start(ctx *telegohandler.Context, message telego.Message) error {
	// _, _, args := telegoutil.ParseCommand(message.Text)
	// if len(args) > 0 {
	// 	common.Logger.Debugf("start: args=%v", args)
	// 	action := strings.Split(args[0], "_")[0]
	// 	switch action {
	// 	case "file":
	// 		pictureID := args[0][5:]
	// 		_, err := utils.SendPictureFileByID(ctx, message, ChannelChatID, pictureID)
	// 		if err != nil {
	// 			utils.ReplyMessage(ctx, ctx.Bot(), message, "获取失败: "+err.Error())
	// 		}
	// 	case "files":
	// 		artworkID := args[0][6:]
	// 		objectID, err := primitive.ObjectIDFromHex(artworkID)
	// 		if err != nil {
	// 			utils.ReplyMessage(ctx, ctx.Bot(), message, "无效的ID")
	// 			return nil
	// 		}
	// 		artwork, err := service.GetArtworkByID(ctx, objectID)
	// 		if err != nil {
	// 			utils.ReplyMessage(ctx, ctx.Bot(), message, "获取失败: "+err.Error())
	// 			return nil
	// 		}
	// 		getArtworkFiles(ctx, ctx.Bot(), message, artwork)
	// 	case "code":
	// 		userID := message.From.ID
	// 		userModel, _ := service.GetUserByTelegramID(ctx, userID)
	// 		if userModel != nil {
	// 			ctx.Bot().SendMessage(ctx, telegoutil.Messagef(message.Chat.ChatID(), "您的此 Telegram 账号 ( %d ) 已经绑定了 ManyACG 账号 %s", userID, userModel.Username))
	// 			return nil
	// 		}
	// 		unauthUserID := args[0][5:]
	// 		objectID, err := primitive.ObjectIDFromHex(unauthUserID)
	// 		if err != nil {
	// 			utils.ReplyMessage(ctx, ctx.Bot(), message, "无效的ID")
	// 			return nil
	// 		}
	// 		unauthUser, err := service.GetUnauthUserByID(ctx, objectID)
	// 		if err != nil {
	// 			utils.ReplyMessage(ctx, ctx.Bot(), message, "获取失败: "+err.Error())
	// 			return nil
	// 		}
	// 		_, err = ctx.Bot().SendMessage(ctx, telegoutil.Messagef(message.Chat.ChatID(),
	// 			"您的此 Telegram 账号 ( %d ) 将与 ManyACG 账号 %s 绑定\n验证码: <code>%s</code>",
	// 			userID,
	// 			common.EscapeHTML(unauthUser.Username),
	// 			common.EscapeHTML(unauthUser.Code)).
	// 			WithParseMode(telego.ModeHTML),
	// 		)
	// 		if err != nil {
	// 			common.Logger.Errorf("Failed to send message: %v", err)
	// 			return nil
	// 		}
	// 		unauthUser.TelegramID = userID
	// 		service.UpdateUnauthUser(ctx, objectID, unauthUser)
	// 	case "info":
	// 		dataID := args[0][5:]
	// 		sourceURL, err := service.GetCallbackDataByID(ctx, dataID)
	// 		if err != nil {
	// 			utils.ReplyMessage(ctx, ctx.Bot(), message, "获取失败: "+err.Error())
	// 			return nil
	// 		}
	// 		if err := utils.SendFullArtworkInfo(ctx, ctx.Bot(), message, sourceURL); err != nil {
	// 			utils.ReplyMessage(ctx, ctx.Bot(), message, err.Error())
	// 		}
	// 	}
	// 	return nil
	// }
	return Help(ctx, message)
}
