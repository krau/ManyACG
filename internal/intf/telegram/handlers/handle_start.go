package handlers

import (
	"fmt"
	"strings"

	"github.com/krau/ManyACG/internal/app/query"
	"github.com/krau/ManyACG/internal/constant/version"
	"github.com/krau/ManyACG/internal/intf/telegram/utils"
	"github.com/krau/ManyACG/internal/pkg/log"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
)

func (h *BotHandlers) Start(ctx *telegohandler.Context, message telego.Message) error {
	_, _, args := telegoutil.ParseCommand(message.Text)
	if len(args) > 0 {
		log.Debug("start", "args", args)
		action := strings.Split(args[0], "_")[0]
		switch action {
		case "file":
			// pictureID := args[0][5:]
			// _, err := utils.SendPictureFileByID(ctx, ctx.Bot(), message, meta.ChannelChatID, pictureID)
			// if err != nil {
			// 	utils.ReplyMessage(ctx, ctx.Bot(), message, "获取失败: "+err.Error())
			// }
		case "files":
			// artworkID := args[0][6:]
			// objectID, err := primitive.ObjectIDFromHex(artworkID)
			// if err != nil {
			// 	utils.ReplyMessage(ctx, ctx.Bot(), message, "无效的ID")
			// 	return nil
			// }
			// artwork, err := service.GetArtworkByID(ctx, objectID)
			// if err != nil {
			// 	utils.ReplyMessage(ctx, ctx.Bot(), message, "获取失败: "+err.Error())
			// 	return nil
			// }
			// getArtworkFiles(ctx, ctx.Bot(), message, artwork)
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
			// 	strutil.EscapeHTML(unauthUser.Username),
			// 	strutil.EscapeHTML(unauthUser.Code)).
			// 	WithParseMode(telego.ModeHTML),
			// )
			// if err != nil {
			// 	common.Logger.Errorf("Failed to send message: %v", err)
			// 	return nil
			// }
			// unauthUser.TelegramID = userID
			// service.UpdateUnauthUser(ctx, objectID, unauthUser)
		case "info":
			// dataID := args[0][5:]
			// sourceURL, err := service.GetCallbackDataByID(ctx, dataID)
			// if err != nil {
			// 	utils.ReplyMessage(ctx, ctx.Bot(), message, "获取失败: "+err.Error())
			// 	return nil
			// }
			// if err := utils.SendFullArtworkInfo(ctx, ctx.Bot(), message, sourceURL); err != nil {
			// 	utils.ReplyMessage(ctx, ctx.Bot(), message, err.Error())
			// }
		}
		return nil
	}
	return h.Help(ctx, message)
}

func (h *BotHandlers) Help(ctx *telegohandler.Context, message telego.Message) error {
	helpText := `使用方法:
/setu - 随机图片(NSFW)
/random - 随机全年龄图片
/search - 搜索相似图片
/info - 发送作品图片和信息
/hash - 计算图片信息
/stats - 获取统计数据
/files - 获取作品原图
/hybrid - 混合搜索作品
/similar - 搜索相似作品
`
	helpText += `
随机图片相关功能中支持使用以下格式的参数:
使用 '|' 分隔'或'关系, 使用 '空格' 分隔'与'关系, 示例:

/random 萝莉|白丝 猫耳|原创

表示搜索包含"萝莉"或"白丝", 且包含"猫耳"或"原创"的图片.
Inline 查询(在任意聊天框中@本bot)支持同样的参数格式.
`
	isAdmin, _ := h.app.Queries.AdminQuery.Handle(ctx, *query.NewAdminQuery(message.From.ID))
	if isAdmin {
		helpText += `
	管理员命令:
	/set_admin - 设置|删除管理员
	/delete - 删除整个作品
	/r18 - 设置作品R18标记
	/title - 设置作品标题
	/tags - 更新作品标签(覆盖原有标签)
	/autotag - 自动tag作品
	/addtags - 添加作品标签
	/deltags - 删除作品标签
	/tagalias - 为标签添加别名
	/dump - 输出 json 格式作品信息
	/recaption - 重新生成作品描述
	`
	}
	helpText += fmt.Sprintf("\n版本: %s, 构建日期 %s, 提交 %s\nhttps://github.com/krau/ManyACG", version.Version, version.BuildTime, version.Commit[:7])
	utils.ReplyMessage(ctx, ctx.Bot(), message, helpText)
	return nil
}
