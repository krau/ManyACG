package handlers

import (
	"ManyACG/common"
	"ManyACG/service"
	"ManyACG/telegram"
	"ManyACG/types"
	"context"
	"math/rand"

	. "ManyACG/logger"

	"github.com/google/uuid"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

func InlineQuery(ctx context.Context, bot *telego.Bot, query telego.InlineQuery) {
	queryText := query.Query
	texts, err := common.ParseStringTo2DArray(queryText, "|", " ")
	if err != nil {
		Logger.Warnf("解析参数失败: %s", err)
		bot.AnswerInlineQuery(telegoutil.InlineQuery(query.ID, telegoutil.ResultArticle(uuid.NewString(), "解析参数失败", telegoutil.TextMessage("/setu"))))
		return
	}
	artworks, err := service.QueryArtworksByTexts(ctx, texts, types.R18TypeAll, 20)
	if err != nil || len(artworks) == 0 {
		Logger.Warnf("获取图片失败: %s", err)
		bot.AnswerInlineQuery(telegoutil.InlineQuery(query.ID, telegoutil.ResultArticle(uuid.NewString(), "未找到相关图片", telegoutil.TextMessage("/setu"))))
		return
	}
	results := make([]telego.InlineQueryResult, 0, len(artworks))
	for _, artwork := range artworks {
		picture := artwork.Pictures[rand.Intn(len(artwork.Pictures))]
		if picture.TelegramInfo.PhotoFileID == "" {
			continue
		}
		result := telegoutil.ResultCachedPhoto(uuid.NewString(), picture.TelegramInfo.PhotoFileID).WithCaption(artwork.Title)
		result.WithReplyMarkup(telegoutil.InlineKeyboard(telegram.GetPostedPictureInlineKeyboardButton(picture)))
		results = append(results, result)
	}
	if err := bot.AnswerInlineQuery(telegoutil.InlineQuery(query.ID, results...).WithCacheTime(1)); err != nil {
		Logger.Errorf("回复查询失败: %s", err)
	}
}
