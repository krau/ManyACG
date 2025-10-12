package handlers

import (
	"context"
	"errors"
	"fmt"
	"html"
	"math/rand"
	"strconv"
	"strings"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/interface/telegram/handlers/utils"
	"github.com/krau/ManyACG/internal/interface/telegram/metautil"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/model/query"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/internal/shared/errs"
	"github.com/krau/ManyACG/pkg/strutil"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/samber/oops"
)

func RandomPicture(ctx *telegohandler.Context, message telego.Message) error {
	cmd, _, args := telegoutil.ParseCommand(message.Text)
	argText := strings.ReplaceAll(strings.Join(args, " "), "\\", "")
	textArray := strutil.ParseTo2DArray(argText, "|", " ")
	r18 := cmd == "setu"
	r18Type := shared.R18TypeNone
	if r18 {
		r18Type = shared.R18TypeR18
	}
	serv := service.FromContext(ctx)
	artwork, err := serv.QueryArtworks(ctx, query.ArtworksDB{
		ArtworksFilter: query.ArtworksFilter{
			R18:      r18Type,
			Keywords: textArray,
		},
		Paginate: query.Paginate{ 
			Offset: 0,
			Limit:  1,
		},
		Random: true,
	})
	if err != nil {
		if errors.Is(err, errs.ErrRecordNotFound) {
			utils.ReplyMessage(ctx, message, "未找到相关图片")
			return nil
		}
		utils.ReplyMessage(ctx, message, "查询图片失败")
		return oops.Wrapf(err, "failed to query artworks for random picture")
	}
	if len(artwork) == 0 {
		utils.ReplyMessage(ctx, message, "未找到相关图片")
		return nil
	}
	pictures := artwork[0].Pictures
	picIndex := rand.Intn(len(pictures))
	picture := pictures[picIndex]
	file, err := utils.GetPicturePhotoInputFile(ctx, serv, picture)
	if err != nil {
		utils.ReplyMessage(ctx, message, "获取图片失败")
		return oops.Wrapf(err, "failed to get picture input file")
	}
	defer file.Close()
	aw := artwork[0]
	photo := telegoutil.
		Photo(message.Chat.ChatID(), file.Value).
		WithCaption(fmt.Sprintf("<a href=\"%s\">%s</a>", aw.SourceURL, html.EscapeString(aw.Title))).
		WithParseMode(telego.ModeHTML).
		WithReplyParameters(&telego.ReplyParameters{
			MessageID: message.MessageID,
		}).
		WithReplyMarkup(telegoutil.InlineKeyboard(utils.GetPostedArtworkInlineKeyboardButton(aw, metautil.FromContext(ctx))))
	if aw.R18 {
		photo.WithHasSpoiler()
	}
	photoMessage, err := ctx.Bot().SendPhoto(ctx, photo)
	if err != nil {
		utils.ReplyMessage(ctx, message, "发送图片失败")
		return oops.Wrapf(err, "failed to send photo message")
	}
	if photoMessage != nil {
		fileId := photoMessage.Photo[len(photoMessage.Photo)-1].FileID
		tginfo := picture.TelegramInfo.Data()
		tginfo.PhotoFileID = fileId
		if err := serv.UpdatePictureTelegramInfo(ctx, picture.ID, &tginfo); err != nil {
			return oops.Wrapf(err, "failed to update picture telegram info")
		}
	}
	return nil
}

func HybridSearchArtworks(ctx *telegohandler.Context, message telego.Message) error {
	serv := service.FromContext(ctx)
	_, _, args := telegoutil.ParseCommand(message.Text)
	if len(args) == 0 {
		helpText := `
<b>使用 /hybrid 命令并提供查询参数, 将使用混合搜索引擎搜索相关图片</b>

命令语法: /hybrid <搜索内容> [语义比例]

语义比例为0-1的浮点数, 应位于参数列表最后, 越大越趋向于基于语义搜索, 若不提供, 使用默认值0.8

<i>Tips: 该命令将基于文本语义进行搜索, 而非关键词匹配</i>
`
		utils.ReplyMessageWithHTML(ctx, message, helpText)
		return nil
	}
	var hybridSemanticRatio float64
	var queryText string
	hybridSemanticRatio, err := strconv.ParseFloat(args[len(args)-1], 64)
	if err != nil {
		hybridSemanticRatio = 0.8
		queryText = strings.Join(args, " ")
	} else {
		if hybridSemanticRatio < 0 || hybridSemanticRatio > 1 {
			utils.ReplyMessage(ctx, message, "参数错误: 语义比例应为0-1的小数")
			return nil
		}
		queryText = strings.Join(args[:len(args)-1], " ")
	}
	artworks, err := serv.SearchArtworks(ctx, &query.ArtworkSearch{
		//  queryText, hybridSemanticRatio, 0, 50, types.R18TypeAll
		Query:               queryText,
		Hybrid:              true,
		HybridSemanticRatio: hybridSemanticRatio,
		Paginate: query.Paginate{
			Offset: 0,
			Limit:  50,
		},
		R18: shared.R18TypeAll,
	})
	if err != nil {
		if errors.Is(err, errs.ErrSearchEngineNotEnabled) {
			utils.ReplyMessage(ctx, message, "搜索引擎未启用")
			return nil
		}
		utils.ReplyMessage(ctx, message, "搜索失败")
		return oops.Wrapf(err, "failed to search artworks")
	}
	if len(artworks) == 0 {
		utils.ReplyMessage(ctx, message, "未找到相关图片")
		return nil
	}

	if len(artworks) > 10 {
		artworks = slice.Shuffle(artworks)[:10]
	}
	return handleSendResultArtworks(ctx, artworks, message, ctx.Bot())
}

func SearchSimilarArtworks(ctx *telegohandler.Context, message telego.Message) error {
	if message.ReplyToMessage == nil {
		helpText := `
<b>使用 /similar 命令回复一条包含图片或作品链接的消息, 将搜索与该图片相关的作品</b>

命令语法: /similar [偏移量] [限制数量]

若回复的消息中未找到支持的链接, 将尝试识别图片内容并搜索相关作品
`
		utils.ReplyMessageWithHTML(ctx, message, helpText)
		return nil
	}
	serv := service.FromContext(ctx)
	// var sourceURL string
	sourceURL := utils.FindSourceURLInMessage(serv, message.ReplyToMessage)
	if sourceURL == "" {
		utils.ReplyMessage(ctx, message, "回复的消息中未找到支持的链接")
	}
	artwork, err := serv.GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		utils.ReplyMessage(ctx, message, "获取作品信息失败")
		return oops.Wrapf(err, "failed to get artwork by url")
	}
	_, _, args := telegoutil.ParseCommand(message.Text)
	offset := 0
	limit := 50
	if len(args) > 0 {
		offset, err = strconv.Atoi(args[0])
		if err != nil || offset < 0 {
			utils.ReplyMessage(ctx, message, "参数错误: 偏移量应为非负整数")
			return nil
		}
	}
	if len(args) > 1 {
		limit, err = strconv.Atoi(args[1])
		if err != nil || limit < 1 || limit > 100 {
			utils.ReplyMessage(ctx, message, "参数错误: 限制数量应为1-10的整数")
			return nil
		}
	}
	artworks, err := serv.FindSimilarArtworks(ctx, &query.ArtworkSimilar{
		ArtworkID: artwork.ID,
		R18:       shared.R18TypeAll,
		Paginate: query.Paginate{
			Offset: offset,
			Limit:  limit,
		},
	})
	if err != nil {
		if errors.Is(err, errs.ErrSearchEngineNotEnabled) {
			utils.ReplyMessage(ctx, message, "搜索引擎未启用")
			return nil
		}
		utils.ReplyMessage(ctx, message, "搜索失败")
		return oops.Wrapf(err, "failed to find similar artworks")
	}
	if len(artworks) == 0 {
		utils.ReplyMessage(ctx, message, "未找到相似的作品")
		return nil
	}
	if len(artworks) > 10 {
		artworks = slice.Shuffle(artworks)[:10]
	}
	return handleSendResultArtworks(ctx, artworks, message, ctx.Bot())
}

// func handleGetSourceURLFromPicture(ctx *telegohandler.Context, serv *service.Service, message telego.Message) (string, []byte, error) {
// 	file, err := utils.GetMessagePhotoFile(ctx, message.ReplyToMessage)
// 	if err != nil {
// 		return "", nil, err
// 	}
// 	hash, err := imgtool.GetImagePhashFromReader(bytes.NewReader(file))
// 	if err != nil {
// 		return "", file, err
// 	}
// 	pictures, err := serv.QueryPicturesByPhash(ctx, query.PicturesPhash{Input: hash, Distance: 10, Limit: 1})
// 	if err != nil {
// 		return "", file, err
// 	}
// 	if len(pictures) == 0 {
// 		return "", file, errors.New("not found similar pictures by hash")
// 	}
// 	return pictures[0].Artwork.SourceURL, file, nil
// }

func handleSendResultArtworks(ctx context.Context, artworks []*entity.Artwork, message telego.Message, bot *telego.Bot) error {
	inputMedias := make([]telego.InputMedia, 0, len(artworks))
	for _, artwork := range artworks {
		picture := artwork.Pictures[0]
		var file telego.InputFile
		if picture.TelegramInfo.Data().PhotoFileID != "" {
			file = telegoutil.FileFromID(picture.TelegramInfo.Data().PhotoFileID)
		} else {
			photoURL := fmt.Sprintf("%s/?url=%s&w=2560&h=2560&we&output=jpg", runtimecfg.Get().Wsrv.URL, picture.Original)
			file = telegoutil.FileFromURL(photoURL)
		}
		caption := fmt.Sprintf("<a href=\"%s\">%s</a>", artwork.SourceURL, html.EscapeString(artwork.Title))
		inputMedias = append(inputMedias, telegoutil.MediaPhoto(file).WithCaption(caption).WithParseMode(telego.ModeHTML))
	}
	mediaGroup := telegoutil.MediaGroup(message.Chat.ChatID(), inputMedias...).WithReplyParameters(&telego.ReplyParameters{
		MessageID: message.MessageID,
		ChatID:    message.Chat.ChatID(),
	})
	_, err := bot.SendMediaGroup(ctx, mediaGroup)
	return err
}
