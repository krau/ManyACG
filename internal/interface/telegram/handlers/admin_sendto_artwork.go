package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/krau/ManyACG/internal/infra/kvstor"
	"github.com/krau/ManyACG/internal/interface/telegram/handlers/utils"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/samber/oops"
)

func SendtoArtworkCallbackQuery(ctx *telegohandler.Context, query telego.CallbackQuery) error {
	serv, err := requireService(ctx)
	if err != nil {
		return err
	}
	answerQuery := func(text string, showAlert bool) {
		ctx.Bot().AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            text,
			ShowAlert:       showAlert,
			CacheTime:       60,
		})
	}
	if !utils.CheckPermissionForQuery(ctx, serv, query, shared.PermissionPostArtwork) {
		answerQuery("你没有发布作品的权限", true)
		return nil
	}
	queryDataSlice := strings.Split(query.Data, " ")
	// sendto <cbId> <chatId>
	if len(queryDataSlice) != 3 {
		answerQuery("无效的回调数据", true)
		return nil
	}
	dataID := queryDataSlice[1]
	chatIdStr := queryDataSlice[2]
	chatId, err := strconv.ParseInt(chatIdStr, 10, 64)
	if err != nil {
		answerQuery("无效的聊天ID", true)
		return nil
	}
	sourceURL, err := kvstor.Get[string](ctx, dataID)
	if err != nil {
		answerQuery("获取回调数据失败", true)
		return nil
	}
	cachedArtwork, err := serv.GetOrFetchCachedArtwork(ctx, sourceURL)
	if err != nil {
		answerQuery("获取作品信息失败: "+err.Error(), true)
		return nil
	}

	ctx.Bot().EditMessageReplyMarkup(ctx,
		telegoutil.EditMessageReplyMarkup(query.Message.GetChat().ChatID(),
			query.Message.GetMessageID(),
			telegoutil.InlineKeyboard([]telego.InlineKeyboardButton{
				telegoutil.InlineKeyboardButton("正在发送").WithCallbackData("noop"),
			})))

	meta, err := requireMeta(ctx)
	if err != nil {
		return err
	}
	results, err := utils.SendArtworkMediaGroup(ctx, ctx.Bot(), serv, meta, telegoutil.ID(chatId), cachedArtwork)
	if err != nil {
		ctx.Bot().SendMessage(ctx, telegoutil.Message(query.Message.GetChat().ChatID(),
			fmt.Sprintf("发送作品到聊天 %d 失败: %s", chatId, err.Error()),
		).WithReplyParameters(&telego.ReplyParameters{
			MessageID: query.Message.GetMessageID(),
		}))
		return err
	}
	data := cachedArtwork.Artwork.Data()
	if err := utils.UpdateCachedArtworkFileID(ctx, results, serv, meta, data); err != nil {
		return oops.Wrapf(err, "failed to update cached artwork after send media group")
	}
	ctx.Bot().EditMessageReplyMarkup(ctx,
		telegoutil.EditMessageReplyMarkup(query.Message.GetChat().ChatID(),
			query.Message.GetMessageID(),
			telegoutil.InlineKeyboard([]telego.InlineKeyboardButton{
				telegoutil.InlineKeyboardButton(fmt.Sprintf(
					"已发送到聊天 %d", chatId,
				)).WithCallbackData("noop")}),
		))
	return nil

}
