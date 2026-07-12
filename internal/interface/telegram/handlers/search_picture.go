package handlers

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/krau/ManyACG/internal/infra/tagging"
	"github.com/krau/ManyACG/internal/interface/telegram/handlers/utils"
	"github.com/krau/ManyACG/internal/interface/telegram/metautil"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/samber/oops"
)

func SearchPicture(ctx *telegohandler.Context, message telego.Message) error {
	serv, err := requireService(ctx)
	if err != nil {
		return err
	}
	meta, err := requireMeta(ctx)
	if err != nil {
		return err
	}
	if message.ReplyToMessage == nil {
		helpText := `
<b>使用 /search 命令回复一条图片消息以搜索图片来源</b>
`
		utils.ReplyMessageWithHTML(ctx, message, helpText)
		return nil
	}
	msg, err := utils.ReplyMessage(ctx, message, "少女祈祷中...")
	if err != nil {
		return oops.Wrapf(err, "reply message failed")
	}
	file, err := utils.GetMessagePhotoFile(ctx, message.ReplyToMessage)
	if err != nil {
		ctx.Bot().EditMessageText(ctx, &telego.EditMessageTextParams{
			ChatID:    msg.Chat.ChatID(),
			MessageID: msg.GetMessageID(),
			Text:      "获取图片文件失败: " + err.Error(),
		})
		return nil
	}
	text, dbExists, err := getDBSearchResultText(ctx, serv, meta, file)
	if err != nil {
		log.Errorf("search in db failed: %s", err)
		ctx.Bot().EditMessageText(ctx, &telego.EditMessageTextParams{
			ChatID:    msg.Chat.ChatID(),
			MessageID: msg.GetMessageID(),
			Text:      "搜索失败",
		})
		return nil
	}
	if dbExists {
		ctx.Bot().EditMessageText(ctx, &telego.EditMessageTextParams{
			ChatID:    msg.Chat.ChatID(),
			MessageID: msg.GetMessageID(),
			Text:      text,
			ParseMode: telego.ModeHTML,
		})
		return nil
	}

	ctx.Bot().EditMessageText(ctx, &telego.EditMessageTextParams{
		ChatID:    msg.Chat.ChatID(),
		MessageID: msg.GetMessageID(),
		Text:      "未在数据库中找到相似图片",
	})
	return nil
}

func getDBSearchResultText(ctx context.Context, serv *service.Service, meta *metautil.MetaData, file []byte) (string, bool, error) {
	hits, err := serv.SearchPicturesByImage(ctx, file, 10, 10)
	if err != nil {
		return "", false, oops.Wrapf(err, "search pictures by image failed")
	}
	shown := 0
	var body strings.Builder
	for _, hit := range hits {
		picture := hit.Picture
		if picture == nil || picture.Artwork == nil {
			continue
		}
		shown++
		body.WriteString(fmt.Sprintf("<a href=\"%s\">%s</a> · 第 %d 张\n",
			picture.Artwork.GetSourceURL(),
			utils.EscapeHTML(picture.Artwork.GetTitle()),
			picture.OrderIndex+1,
		))
		if meta.ChannelAvailable() && picture.TelegramInfo.Data().MessageID(meta.ChannelChatID().ID) != 0 {
			body.WriteString(fmt.Sprintf("<a href=\"%s\">频道消息</a>\n", meta.ChannelMessageURL(picture.TelegramInfo.Data().MessageID(meta.ChannelChatID().ID))))
		}
		if meta.SiteURL() != "" {
			body.WriteString(fmt.Sprintf("<a href=\"%s\">网站页面</a>\n", meta.SiteURL()+"/artwork/"+picture.ArtworkID.Hex()))
		}
		body.WriteString("\n")
	}
	if shown == 0 {
		return "未在数据库中找到相似图片", false, nil
	}
	var text strings.Builder
	text.WriteString(fmt.Sprintf("找到 %d 张相似图片\n\n", shown))
	text.WriteString(body.String())
	return text.String(), true, nil
}

func SearchPictureCallbackQuery(ctx *telegohandler.Context, query telego.CallbackQuery) error {
	serv, err := requireService(ctx)
	if err != nil {
		return err
	}
	meta, err := requireMeta(ctx)
	if err != nil {
		return err
	}
	if !query.Message.IsAccessible() {
		return nil
	}
	message, ok := query.Message.(*telego.Message)
	if !ok {
		return oops.Errorf("unexpected message type: %T", query.Message)
	}
	file, err := utils.GetMessagePhotoFile(ctx, message)
	if err != nil {
		ctx.Bot().AnswerCallbackQuery(ctx, telegoutil.CallbackQuery(query.ID).WithText("获取图片文件失败: "+err.Error()).WithShowAlert().WithCacheTime(5))
		return nil
	}
	text, hasResult, err := getDBSearchResultText(ctx, serv, meta, file)
	if err != nil {
		ctx.Bot().AnswerCallbackQuery(ctx, telegoutil.CallbackQuery(query.ID).WithText(err.Error()).WithShowAlert().WithCacheTime(5))
		return nil
	}
	if !hasResult {
		go ctx.Bot().AnswerCallbackQuery(ctx, telegoutil.CallbackQuery(query.ID).WithText(text).WithCacheTime(5))
	} else {
		go ctx.Bot().AnswerCallbackQuery(ctx, telegoutil.CallbackQuery(query.ID).WithText("搜索到相似图片").WithCacheTime(5))
	}
	utils.ReplyMessageWithHTML(ctx, *message, text)
	return nil
}

// type ascii2dResult struct {
// 	Name      string
// 	Link      string
// 	Thumbnail string
// }

// const (
// 	ascii2dAPI = "https://ascii2d.net/search/file"
// 	ascii2dURL = "https://ascii2d.net"
// )

// func getAscii2dSearchResult(file []byte) ([]*ascii2dResult, error) {
// 	respcolor, err := common.Client.R().SetFileBytes("file", "image.jpg", file).Post(ascii2dAPI)
// 	if err != nil {
// 		return nil, fmt.Errorf("请求 ascii2d 失败: %w", err)
// 	}
// 	if respcolor.IsErrorState() {
// 		return nil, fmt.Errorf("请求 ascii2d 失败: %s", respcolor.Status)
// 	}
// 	bovwUrl := func() string {
// 		if respcolor.Response != nil && respcolor.Response.Request != nil && respcolor.Response.Request.Response != nil && respcolor.Response.Request.Response.Header != nil {
// 			return respcolor.Response.Request.Response.Header.Get("Location")
// 		}
// 		return ""
// 	}()
// 	if bovwUrl == "" {
// 		return nil, errors.New("无法获取 bovw 页面")
// 	}
// 	common.Logger.Debugf("getting ascii2d bovw url: %s", bovwUrl)

// 	respbovw, err := common.Client.R().Get(bovwUrl)
// 	if err != nil {
// 		return nil, fmt.Errorf("请求 ascii2d bovw 页面失败: %w", err)
// 	}
// 	if respbovw.IsErrorState() {
// 		return nil, fmt.Errorf("请求 ascii2d bovw 页面失败: %s", respbovw.Status)
// 	}

// 	results := make([]*ascii2dResult, 0)
// 	doc, err := goquery.NewDocumentFromReader(respbovw.Body)
// 	if err != nil {
// 		return nil, fmt.Errorf("解析 ascii2d 页面失败: %w", err)
// 	}

// 	doc.Find(".row.item-box").Each(func(i int, s *goquery.Selection) {
// 		if i >= 10 {
// 			return
// 		}

// 		detail := s.Find(".detail-box h6")
// 		name := detail.First().Find("a").First().Text()
// 		link, exists := detail.Find("a").First().Attr("href")
// 		if !exists {
// 			return
// 		}
// 		thumbnail, exists := s.Find(".image-box img").Attr("src")
// 		if !exists {
// 			return
// 		}
// 		thumbnail = ascii2dURL + thumbnail

// 		results = append(results, &ascii2dResult{
// 			Name:      strings.TrimSpace(name),
// 			Link:      link,
// 			Thumbnail: thumbnail,
// 		})
// 	})

// 	return results, nil

// }

func TaggingPicture(ctx *telegohandler.Context, message telego.Message) error {
	if !tagging.Enabled() {
		utils.ReplyMessage(ctx, message, "标签识别服务未启用")
		return nil
	}
	if message.ReplyToMessage == nil {
		helpText := `
<b>使用 /tagging 命令回复一条图片消息以识别图片中的标签</b>
`
		utils.ReplyMessageWithHTML(ctx, message, helpText)
		return nil
	}

	msg, err := utils.ReplyMessage(ctx, message, "少女祈祷中...")
	if err != nil {
		return oops.Wrapf(err, "reply message failed")
	}
	file, err := utils.GetMessagePhotoFile(ctx, message.ReplyToMessage)
	if err != nil {
		ctx.Bot().EditMessageText(ctx, &telego.EditMessageTextParams{
			ChatID:    msg.Chat.ChatID(),
			MessageID: msg.GetMessageID(),
			Text:      "获取图片文件失败: " + err.Error(),
		})
		return nil
	}

	serv, err := requireService(ctx)
	if err != nil {
		return err
	}
	result, err := serv.Tagger().Predict(ctx, bytes.NewReader(file))
	if err != nil {
		log.Errorf("tagging predict failed: %s", err)
		ctx.Bot().EditMessageText(ctx, &telego.EditMessageTextParams{
			ChatID:    msg.Chat.ChatID(),
			MessageID: msg.GetMessageID(),
			Text:      "标签识别失败: " + err.Error(),
		})
		return nil
	}

	tags := make([]string, 0, len(result))
	for tag := range result {
		tags = append(tags, tag)
	}
	sort.Strings(tags)

	var sb strings.Builder
	sb.WriteString("<pre>")
	for i, tag := range tags {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(tag)
	}
	sb.WriteString("</pre>")

	ctx.Bot().EditMessageText(ctx, &telego.EditMessageTextParams{
		ChatID:    msg.Chat.ChatID(),
		MessageID: msg.GetMessageID(),
		Text:      sb.String(),
		ParseMode: telego.ModeHTML,
	})
	return nil
}
