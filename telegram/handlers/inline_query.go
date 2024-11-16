package handlers

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/krau/ManyACG/adapter"
	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/service"
	"github.com/krau/ManyACG/sources"
	"github.com/krau/ManyACG/telegram/utils"
	"github.com/krau/ManyACG/types"

	. "github.com/krau/ManyACG/logger"

	"github.com/google/uuid"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

func InlineQuery(ctx context.Context, bot *telego.Bot, query telego.InlineQuery) {
	queryText := query.Query
	if url := sources.FindSourceURL(queryText); url != "" {
		artwork, err := service.GetArtworkByURLWithCacheFetch(ctx, url)
		if err != nil {
			bot.AnswerInlineQuery(telegoutil.InlineQuery(query.ID, telegoutil.ResultArticle(uuid.NewString(), "获取作品失败", telegoutil.TextMessage(url))))
			return
		}
		results := make([]telego.InlineQueryResult, 0, min(len(artwork.Pictures), 48))
		caption := utils.GetArtworkHTMLCaption(artwork)
		for _, picture := range artwork.Pictures {
			if picture.TelegramInfo != nil && picture.TelegramInfo.PhotoFileID != "" {
				result := telegoutil.ResultCachedPhoto(uuid.NewString(), picture.TelegramInfo.PhotoFileID).WithCaption(caption).WithParseMode(telego.ModeHTML)
				results = append(results, result)
				continue
			}
			result := telegoutil.ResultPhoto(uuid.NewString(),
				fmt.Sprintf("https://wsrv.nl/?url=%s&w=2560&h=2560&we",
					picture.Original), picture.Thumbnail).WithCaption(caption).WithParseMode(telego.ModeHTML)
			results = append(results, result)
		}
		if err := bot.AnswerInlineQuery(&telego.AnswerInlineQueryParams{
			InlineQueryID: query.ID,
			Results:       results,
			CacheTime:     10,
		}); err != nil {
			Logger.Errorf("响应Inline查询失败: %s", err)
		}
		return
	}
	texts := common.ParseStringTo2DArray(queryText, "|", " ")
	artworks, err := service.QueryArtworksByTexts(ctx, texts, types.R18TypeAll, 48, adapter.OnlyLoadPicture())
	if err != nil || len(artworks) == 0 {
		Logger.Warnf("获取图片失败: %s", err)
		bot.AnswerInlineQuery(telegoutil.InlineQuery(query.ID, telegoutil.ResultArticle(uuid.NewString(), "未找到相关图片", telegoutil.TextMessage("/setu"))))
		return
	}
	results := make([]telego.InlineQueryResult, 0, len(artworks))
	for _, artwork := range artworks {
		pictureIndex := rand.Intn(len(artwork.Pictures))
		picture := artwork.Pictures[pictureIndex]
		if picture.TelegramInfo == nil || picture.TelegramInfo.PhotoFileID == "" {
			continue
		}
		result := telegoutil.ResultCachedPhoto(uuid.NewString(), picture.TelegramInfo.PhotoFileID).WithCaption(fmt.Sprintf("<a href=\"%s\">%s</a>", artwork.SourceURL, common.EscapeHTML(artwork.Title))).WithParseMode(telego.ModeHTML)
		result.WithReplyMarkup(telegoutil.InlineKeyboard(utils.GetPostedPictureInlineKeyboardButton(artwork, uint(pictureIndex), ChannelChatID, BotUsername)))
		results = append(results, result)
	}
	if err := bot.AnswerInlineQuery(&telego.AnswerInlineQueryParams{
		InlineQueryID: query.ID,
		Results:       results,
		CacheTime:     1,
	}); err != nil {
		Logger.Errorf("响应Inline查询失败: %s", err)
	}
}
