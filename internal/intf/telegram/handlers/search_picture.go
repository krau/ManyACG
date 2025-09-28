package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/krau/ManyACG/internal/common"
	"github.com/krau/ManyACG/internal/infra/config"
	"github.com/krau/ManyACG/internal/pkg/imgtool"

	"github.com/krau/ManyACG/internal/intf/telegram/utils"
	"github.com/krau/ManyACG/service"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func SearchPicture(ctx *telegohandler.Context, message telego.Message) error {
	if message.ReplyToMessage == nil {
		helpText := `
<b>使用 /search 命令回复一条图片消息, 将搜索图片来源</b>
`
		utils.ReplyMessageWithHTML(ctx, ctx.Bot(), message, helpText)
		return nil
	}
	msg, err := utils.ReplyMessage(ctx, ctx.Bot(), message, "少女祈祷中...")
	if err != nil {
		common.Logger.Errorf("reply message failed: %s", err)
		return nil
	}

	file, err := utils.GetMessagePhotoFile(ctx, ctx.Bot(), message.ReplyToMessage)
	if err != nil {
		ctx.Bot().EditMessageText(ctx, &telego.EditMessageTextParams{
			ChatID:    msg.Chat.ChatID(),
			MessageID: msg.GetMessageID(),
			Text:      "获取图片文件失败: " + err.Error(),
		})
		return nil
	}
	text, dbExists, err := getDBSearchResultText(ctx, file)
	if err != nil {
		common.Logger.Errorf("search in db failed: %s", err)
	}
	if dbExists {
		ctx.Bot().EditMessageText(ctx, &telego.EditMessageTextParams{
			ChatID:    msg.Chat.ChatID(),
			MessageID: msg.GetMessageID(),
			Text:      text,
			ParseMode: telego.ModeMarkdownV2,
		})
		return nil
	} else {
		go ctx.Bot().EditMessageText(ctx, &telego.EditMessageTextParams{
			ChatID:    msg.Chat.ChatID(),
			MessageID: msg.GetMessageID(),
			Text:      "数据库搜索无结果, 使用 ascii2d 搜索中...",
		})
	}

	ascii2dResults, err := getAscii2dSearchResult(file)
	if err != nil {
		common.Logger.Errorf("search in ascii2d failed: %s", err)
		ctx.Bot().EditMessageText(ctx, &telego.EditMessageTextParams{
			ChatID:    msg.Chat.ChatID(),
			MessageID: msg.GetMessageID(),
			Text:      "ascii2d 搜索失败",
		})
		return nil
	}
	if len(ascii2dResults) == 0 {
		ctx.Bot().EditMessageText(ctx, &telego.EditMessageTextParams{
			ChatID:    msg.Chat.ChatID(),
			MessageID: msg.GetMessageID(),
			Text:      "没有搜索到相似图片",
		})
		return nil
	}
	text = fmt.Sprintf("在 ascii2d 搜索到%d张相似的图片\n\n", len(ascii2dResults))
	for _, result := range ascii2dResults {
		text += fmt.Sprintf("[%s](%s)\n\n", common.EscapeMarkdown(result.Name), common.EscapeMarkdown(result.Link))
	}
	thumbFile, err := common.DownloadWithCache(ctx, ascii2dResults[0].Thumbnail, nil)
	if err != nil {
		common.Logger.Errorf("download thumbnail failed: %s", err)
		ctx.Bot().EditMessageText(ctx, &telego.EditMessageTextParams{
			ChatID:    msg.Chat.ChatID(),
			MessageID: msg.GetMessageID(),
			Text:      text,
			ParseMode: telego.ModeMarkdownV2,
		})
	} else {
		_, err = ctx.Bot().EditMessageMedia(ctx, &telego.EditMessageMediaParams{
			ChatID:    msg.Chat.ChatID(),
			MessageID: msg.GetMessageID(),
			Media: telegoutil.MediaPhoto(telegoutil.File(telegoutil.NameReader(bytes.NewReader(thumbFile), path.Base(ascii2dResults[0].Thumbnail)))).
				WithCaption(text).
				WithParseMode(telego.ModeMarkdownV2),
		})
		if err != nil {
			common.Logger.Errorf("edit message media failed: %s", err)
		}
	}
	return nil
}

func SearchPictureCallbackQuery(ctx *telegohandler.Context, query telego.CallbackQuery) error {
	if !query.Message.IsAccessible() {
		return nil
	}
	message := query.Message.(*telego.Message)
	file, err := utils.GetMessagePhotoFile(ctx, ctx.Bot(), message)
	if err != nil {
		ctx.Bot().AnswerCallbackQuery(ctx, telegoutil.CallbackQuery(query.ID).WithText("获取图片文件失败: "+err.Error()).WithShowAlert().WithCacheTime(5))
		return nil
	}
	text, hasResult, err := getDBSearchResultText(ctx, file)
	if err != nil {
		ctx.Bot().AnswerCallbackQuery(ctx, telegoutil.CallbackQuery(query.ID).WithText(err.Error()).WithShowAlert().WithCacheTime(5))
		return nil
	}
	if !hasResult {
		go ctx.Bot().AnswerCallbackQuery(ctx, telegoutil.CallbackQuery(query.ID).WithText(text).WithCacheTime(5))
	} else {
		go ctx.Bot().AnswerCallbackQuery(ctx, telegoutil.CallbackQuery(query.ID).WithText("搜索到相似图片").WithCacheTime(5))
	}
	utils.ReplyMessageWithMarkdown(ctx, ctx.Bot(), *message, text)
	return nil
}

func getDBSearchResultText(ctx context.Context, file []byte) (string, bool, error) {
	hash, err := imgtool.GetImagePhashFromReader(bytes.NewReader(file))
	if err != nil {
		return "", false, fmt.Errorf("获取图片哈希失败: %w", err)
	}
	pictures, err := service.GetPicturesByHashHammingDistance(ctx, hash, 10)
	if err != nil {
		return "", false, fmt.Errorf("搜索图片失败: %w", err)
	}
	channelMessageAvailable := ChannelChatID.ID != 0 || ChannelChatID.Username != ""
	enableSite := config.Get().API.SiteURL != ""
	if len(pictures) == 0 {
		return "未在数据库中找到相似图片", false, nil
	}
	text := fmt.Sprintf("找到%d张相似的图片\n\n", len(pictures))
	for _, picture := range pictures {
		artworkObjectID, err := primitive.ObjectIDFromHex(picture.ArtworkID)
		if err != nil {
			common.Logger.Errorf("无效的ObjectID: %s", picture.ID)
			continue
		}
		artwork, err := service.GetArtworkByID(ctx, artworkObjectID)
		if err != nil {
			common.Logger.Errorf("获取作品信息失败: %s", err)
			continue
		}
		text += fmt.Sprintf("[%s\\_%d](%s)\n",
			common.EscapeMarkdown(artwork.Title),
			picture.Index+1,
			common.EscapeMarkdown(artwork.SourceURL),
		)
		if channelMessageAvailable && picture.TelegramInfo != nil && picture.TelegramInfo.MessageID != 0 {
			text += fmt.Sprintf("[频道消息](%s)\n", utils.GetArtworkPostMessageURL(picture.TelegramInfo.MessageID, ChannelChatID))
		}
		if enableSite {
			text += fmt.Sprintf("[ManyACG](%s)\n\n", config.Get().API.SiteURL+"/artwork/"+artwork.ID)
		}
	}
	return text, true, nil
}

type ascii2dResult struct {
	Name      string
	Link      string
	Thumbnail string
}

const (
	ascii2dAPI = "https://ascii2d.net/search/file"
	ascii2dURL = "https://ascii2d.net"
)

func getAscii2dSearchResult(file []byte) ([]*ascii2dResult, error) {
	respcolor, err := common.Client.R().SetFileBytes("file", "image.jpg", file).Post(ascii2dAPI)
	if err != nil {
		return nil, fmt.Errorf("请求 ascii2d 失败: %w", err)
	}
	if respcolor.IsErrorState() {
		return nil, fmt.Errorf("请求 ascii2d 失败: %s", respcolor.Status)
	}
	bovwUrl := func() string {
		if respcolor.Response != nil && respcolor.Response.Request != nil && respcolor.Response.Request.Response != nil && respcolor.Response.Request.Response.Header != nil {
			return respcolor.Response.Request.Response.Header.Get("Location")
		}
		return ""
	}()
	if bovwUrl == "" {
		return nil, errors.New("无法获取 bovw 页面")
	}
	common.Logger.Debugf("getting ascii2d bovw url: %s", bovwUrl)

	respbovw, err := common.Client.R().Get(bovwUrl)
	if err != nil {
		return nil, fmt.Errorf("请求 ascii2d bovw 页面失败: %w", err)
	}
	if respbovw.IsErrorState() {
		return nil, fmt.Errorf("请求 ascii2d bovw 页面失败: %s", respbovw.Status)
	}

	results := make([]*ascii2dResult, 0)
	doc, err := goquery.NewDocumentFromReader(respbovw.Body)
	if err != nil {
		return nil, fmt.Errorf("解析 ascii2d 页面失败: %w", err)
	}

	doc.Find(".row.item-box").Each(func(i int, s *goquery.Selection) {
		if i >= 10 {
			return
		}

		detail := s.Find(".detail-box h6")
		name := detail.First().Find("a").First().Text()
		link, exists := detail.Find("a").First().Attr("href")
		if !exists {
			return
		}
		thumbnail, exists := s.Find(".image-box img").Attr("src")
		if !exists {
			return
		}
		thumbnail = ascii2dURL + thumbnail

		results = append(results, &ascii2dResult{
			Name:      strings.TrimSpace(name),
			Link:      link,
			Thumbnail: thumbnail,
		})
	})

	return results, nil

}
