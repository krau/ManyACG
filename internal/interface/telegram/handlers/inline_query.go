package handlers

import (
	"fmt"
	"html"
	"math/rand"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/interface/telegram/handlers/utils"
	"github.com/krau/ManyACG/internal/interface/telegram/metautil"
	"github.com/krau/ManyACG/internal/model/query"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"github.com/krau/ManyACG/pkg/strutil"
	"github.com/samber/oops"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
)

func InlineQuery(ctx *telegohandler.Context, inlineQuery telego.InlineQuery) error {
	queryText := inlineQuery.Query
	serv := service.FromContext(ctx)
	meta := metautil.FromContext(ctx)
	if url := serv.FindSourceURL(queryText); url != "" {
		artwork, err := serv.GetOrFetchCachedArtwork(ctx, url)
		if err != nil {
			ctx.Bot().AnswerInlineQuery(ctx, telegoutil.InlineQuery(inlineQuery.ID, telegoutil.ResultArticle(objectuuid.New().Hex(), "获取作品失败", telegoutil.TextMessage(url))))
			return nil
		}
		pics := artwork.Artwork.Data().Pictures
		results := make([]telego.InlineQueryResult, 0, min(len(pics), 48))
		caption := utils.ArtworkHTMLCaption(artwork)
		wsrvUrl := runtimecfg.Get().Wsrv.URL
		for _, picture := range pics {
			if picFileId := picture.TelegramInfo.FileID(meta.BotID(), shared.TelegramMediaTypePhoto); picFileId != "" {
				result := telegoutil.ResultCachedPhoto(objectuuid.New().Hex(), picFileId).WithCaption(caption).WithParseMode(telego.ModeHTML)
				results = append(results, result)
				continue
			}
			if wsrvUrl == "" {
				continue
			}
			result := telegoutil.ResultPhoto(objectuuid.New().Hex(),
				fmt.Sprintf("%s/?url=%s&w=2560&h=2560&we&output=jpg", wsrvUrl,
					picture.Original), picture.Thumbnail).WithCaption(caption).WithParseMode(telego.ModeHTML)
			results = append(results, result)
		}
		if err := ctx.Bot().AnswerInlineQuery(ctx, &telego.AnswerInlineQueryParams{
			InlineQueryID: inlineQuery.ID,
			Results:       results,
			CacheTime:     10,
		}); err != nil {
			return oops.Wrapf(err, "failed to answer inline query")
		}
		return nil
	}
	texts := strutil.ParseTo2DArray(queryText, "|", " ")
	artworks, err := serv.QueryArtworks(ctx, query.ArtworksDB{
		ArtworksFilter: query.ArtworksFilter{
			R18:      shared.R18TypeAll,
			Keywords: texts,
		},
		Paginate: query.Paginate{
			Limit:  48,
			Offset: 0,
		},
		Random: true,
	})
	if err != nil || len(artworks) == 0 {
		log.Errorf("获取图片失败: %s", err)
		ctx.Bot().AnswerInlineQuery(ctx, telegoutil.InlineQuery(inlineQuery.ID, telegoutil.ResultArticle(objectuuid.New().Hex(), "未找到相关图片", telegoutil.TextMessage(fmt.Sprintf(`
未找到相关图片 (搜索: %s)

<b>在任意聊天框中输入 @%s [关键词参数] 来查找相关图片</b>`, html.EscapeString(queryText), html.EscapeString(meta.BotUsername()))).WithParseMode(telego.ModeHTML))))
		return nil
	}
	results := make([]telego.InlineQueryResult, 0, len(artworks))
	for _, artwork := range artworks {
		pictureIndex := rand.Intn(len(artwork.Pictures))
		picture := artwork.Pictures[pictureIndex]
		if picture.TelegramInfo.Data().FileID(meta.BotID(), shared.TelegramMediaTypePhoto) == "" {
			continue
		}
		result := telegoutil.
			ResultCachedPhoto(objectuuid.New().Hex(),
				picture.TelegramInfo.Data().FileID(meta.BotID(), shared.TelegramMediaTypePhoto)).
			WithCaption(fmt.Sprintf("<a href=\"%s\">%s</a>", artwork.SourceURL, html.EscapeString(artwork.Title))).
			WithParseMode(telego.ModeHTML).WithReplyMarkup(telegoutil.InlineKeyboard(utils.GetPostedArtworkInlineKeyboardButton(artwork, meta)))
		results = append(results, result)
	}
	if err := ctx.Bot().AnswerInlineQuery(ctx, &telego.AnswerInlineQueryParams{
		InlineQueryID: inlineQuery.ID,
		Results:       results,
		CacheTime:     1,
	}); err != nil {
		return oops.Wrapf(err, "failed to answer inline query")
	}
	return nil
}
