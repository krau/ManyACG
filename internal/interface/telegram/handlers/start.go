package handlers

import (
	"strings"

	"github.com/krau/ManyACG/internal/interface/telegram/handlers/utils"
	"github.com/krau/ManyACG/internal/interface/telegram/metautil"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"github.com/krau/ManyACG/service"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/samber/oops"
)

func Start(ctx *telegohandler.Context, message telego.Message) error {
	_, _, args := telegoutil.ParseCommand(message.Text)
	if len(args) > 0 {
		log.Debug("received start", "args", args)
		serv := service.FromContext(ctx)
		action := strings.Split(args[0], "_")[0]
		switch action {
		case "file": // for compatibility we keep "file" action to get single picture by id
			pictureIDStr := args[0][5:]
			pictureID, err := objectuuid.FromObjectIDHex(pictureIDStr)
			if err != nil {
				utils.ReplyMessage(ctx, message, "无效的ID")
				return nil
			}
			picture, err := serv.GetPictureByID(ctx, pictureID)
			if err != nil {
				utils.ReplyMessage(ctx, message, "获取失败")
				return oops.Wrapf(err, "failed to get picture by id %s", pictureIDStr)
			}
			file, err := utils.GetPictureDocumentInputFile(ctx, serv, picture.Artwork, picture)
			if err != nil {
				utils.ReplyMessage(ctx, message, "获取失败")
				return oops.Wrapf(err, "failed to get picture document input file by id %s", pictureIDStr)
			}
			defer file.Close()
			_, err = ctx.Bot().SendDocument(ctx, telegoutil.Document(message.Chat.ChatID(), file.InputFile))
			return err
		case "files":
			artworkIDStr := args[0][6:]
			artworkID, err := objectuuid.FromObjectIDHex(artworkIDStr)
			if err != nil {
				utils.ReplyMessage(ctx, message, "无效的ID")
				return nil
			}
			artwork, err := serv.GetArtworkByID(ctx, artworkID)
			if err != nil {
				utils.ReplyMessage(ctx, message, "获取失败")
				return oops.Wrapf(err, "failed to get artwork by id: %s", artworkIDStr)
			}
			return getArtworkFiles(ctx, serv, metautil.FromContext(ctx), message, artwork)
		case "code":
			// userID := message.From.ID
			// userModel, _ := service.GetUserByTelegramID(ctx, userID)
			// if userModel != nil {
			// 	ctx.Bot().SendMessage(ctx, telegoutil.Messagef(message.Chat.ChatID(), "您的此 Telegram 账号 ( %d ) 已经绑定了 ManyACG 账号 %s", userID, userModel.Username))
			// 	return nil
			// }
			// unauthUserID := args[0][5:]
			// objectID, err := primitive.ObjectIDFromHex(unauthUserID)
			// if err != nil {
			// 	utils.ReplyMessage(ctx, ctx.Bot(), message, "无效的ID")
			// 	return nil
			// }
			// unauthUser, err := service.GetUnauthUserByID(ctx, objectID)
			// if err != nil {
			// 	utils.ReplyMessage(ctx, ctx.Bot(), message, "获取失败: "+err.Error())
			// 	return nil
			// }
			// _, err = ctx.Bot().SendMessage(ctx, telegoutil.Messagef(message.Chat.ChatID(),
			// 	"您的此 Telegram 账号 ( %d ) 将与 ManyACG 账号 %s 绑定\n验证码: <code>%s</code>",
			// 	userID,
			// 	common.EscapeHTML(unauthUser.Username),
			// 	common.EscapeHTML(unauthUser.Code)).
			// 	WithParseMode(telego.ModeHTML),
			// )
			// if err != nil {
			// 	common.Logger.Errorf("Failed to send message: %v", err)
			// 	return nil
			// }
			// unauthUser.TelegramID = userID
			// service.UpdateUnauthUser(ctx, objectID, unauthUser)
		case "info":
			dataID := args[0][5:]
			sourceURL, err := serv.GetStringDataByID(ctx, dataID)
			if err != nil {
				utils.ReplyMessage(ctx, message, "获取失败")
				return oops.Wrapf(err, "failed to get string data by id: %s", dataID)
			}
			artwork, err := serv.GetArtworkByURL(ctx, sourceURL)
			if err != nil {
				utils.ReplyMessage(ctx, message, "获取作品信息失败")
				return oops.Wrapf(err, "failed to get artwork by url: %s", sourceURL)
			}
			msgs, err := utils.SendArtworkMediaGroup(ctx, message.Chat.ChatID(), artwork)
			if err != nil {
				utils.ReplyMessage(ctx, message, "发送作品信息失败")
				return oops.Wrapf(err, "failed to send artwork media group")
			}
			if len(msgs) == 0 {
				log.Warn("no messages sent for artwork", "title", artwork.GetTitle(), "url", artwork.GetSourceURL())
				return nil
			}
			if len(artwork.GetPictures()) != len(msgs) {
				log.Warn("number of messages sent does not match number of pictures", "sent", len(msgs), "pictures", len(artwork.GetPictures()), "title", artwork.GetTitle(), "url", artwork.GetSourceURL())
			}
			for i, pic := range artwork.Pictures {
				if photoSize := msgs[i].Photo; len(photoSize) > 0 {
					tginfo := pic.GetTelegramInfo()
					tginfo.PhotoFileID = photoSize[len(photoSize)-1].FileID
					if err := serv.UpdatePictureTelegramInfo(ctx, pic.ID, &tginfo); err != nil {
						log.Warn("failed to update picture telegram info", "err", err, "picture_id", pic.ID)
					}
				}
			}
		}
		return nil
	}
	return Help(ctx, message)
}
