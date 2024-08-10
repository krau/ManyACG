package handlers

import (
	"ManyACG/adapter"
	"ManyACG/common"
	"ManyACG/service"
	"ManyACG/telegram/utils"
	"ManyACG/types"
	"context"
	"fmt"
	"math/rand"

	. "ManyACG/logger"

	"github.com/google/uuid"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

func InlineQuery(ctx context.Context, bot *telego.Bot, query telego.InlineQuery) {
	queryText := query.Query
	texts := common.ParseStringTo2DArray(queryText, "|", " ")
	artworks, err := service.QueryArtworksByTexts(ctx, texts, types.R18TypeAll, 20, adapter.OnlyLoadPicture())
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
		result := telegoutil.ResultCachedPhoto(uuid.NewString(), picture.TelegramInfo.PhotoFileID).WithCaption(fmt.Sprintf("<a href=\"%s\">%s</a>", artwork.SourceURL, common.EscapeHTML(artwork.Title))).WithParseMode(telego.ModeHTML)
		result.WithReplyMarkup(telegoutil.InlineKeyboard(utils.GetPostedPictureInlineKeyboardButton(picture, ChannelChatID, BotUsername)))
		results = append(results, result)
	}
	if err := bot.AnswerInlineQuery(&telego.AnswerInlineQueryParams{
		InlineQueryID: query.ID,
		Results:       results,
		CacheTime:     1,
	}); err != nil {
		Logger.Errorf("回复查询失败: %s", err)
	}
}
