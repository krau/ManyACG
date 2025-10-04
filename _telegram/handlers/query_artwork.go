package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/krau/ManyACG/adapter"
	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/common/imgtool"
	"github.com/krau/ManyACG/config"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/model/query"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/objectuuid"

	"github.com/krau/ManyACG/service"
	"github.com/krau/ManyACG/telegram/utils"
	"github.com/krau/ManyACG/types"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func RandomPicture(ctx *telegohandler.Context, message telego.Message) error {
	cmd, _, args := telegoutil.ParseCommand(message.Text)
	argText := strings.ReplaceAll(strings.Join(args, " "), "\\", "")
	textArray := common.ParseStringTo2DArray(argText, "|", " ")
	r18 := cmd == "setu"
	r18Type := types.R18TypeNone
	if r18 {
		r18Type = types.R18TypeOnly
	}
	artwork, err := service.QueryArtworksByTexts(ctx, textArray, r18Type, 1, adapter.OnlyLoadPicture())
	if err != nil {
		common.Logger.Errorf("获取图片失败: %s", err)
		text := "获取图片失败"
		if errors.Is(err, mongo.ErrNoDocuments) {
			text = "未找到图片"
		}
		utils.ReplyMessage(ctx, ctx.Bot(), message, text)
		return nil
	}
	if len(artwork) == 0 {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "未找到图片")
		return nil
	}
	pictures := artwork[0].Pictures
	picture := pictures[rand.Intn(len(pictures))]
	var file telego.InputFile
	if picture.TelegramInfo.PhotoFileID != "" {
		file = telegoutil.FileFromID(picture.TelegramInfo.PhotoFileID)
	} else {
		photoURL := fmt.Sprintf("%s/?url=%s&w=2560&h=2560&we&output=jpg", config.Cfg.WSRVURL, picture.Original)
		file = telegoutil.FileFromURL(photoURL)
	}
	caption := fmt.Sprintf("[%s](%s)", common.EscapeMarkdown(artwork[0].Title), artwork[0].SourceURL)
	photo := telegoutil.Photo(message.Chat.ChatID(), file).
		WithReplyParameters(&telego.ReplyParameters{
			MessageID: message.MessageID,
		}).WithCaption(caption).WithParseMode(telego.ModeMarkdownV2).WithReplyMarkup(
		telegoutil.InlineKeyboard(utils.GetPostedPictureInlineKeyboardButton(artwork[0], 0, ChannelChatID, BotUsername)),
	)
	if artwork[0].R18 {
		photo.WithHasSpoiler()
	}
	photoMessage, err := ctx.Bot().SendPhoto(ctx, photo)
	if err != nil {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "发送图片失败: "+err.Error())
	}
	if photoMessage != nil {
		picture.TelegramInfo.PhotoFileID = photoMessage.Photo[len(photoMessage.Photo)-1].FileID
		if service.UpdatePictureTelegramInfo(ctx, picture, picture.TelegramInfo) != nil {
			common.Logger.Warnf("更新图片信息失败: %s", err)
		}
	}
	return nil
}

func HybridSearchArtworks(ctx *telegohandler.Context, message telego.Message) error {
	if common.MeilisearchClient == nil {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "未启用混合搜索功能")
		return nil
	}
	_, _, args := telegoutil.ParseCommand(message.Text)
	if len(args) == 0 {
		helpText := fmt.Sprintf(`
<b>使用 /hybrid 命令并提供查询参数, 将使用混合搜索引擎搜索相关图片</b>

命令语法: %s

语义比例为0-1的浮点数, 应位于参数列表最后, 越大越趋向于基于语义搜索, 若不提供, 使用默认值0.8

<i>Tips: 该命令将基于文本语义进行搜索, 而非关键词匹配</i>
`, common.EscapeHTML("/hybrid <搜索内容> [语义比例]"))
		utils.ReplyMessageWithHTML(ctx, ctx.Bot(), message, helpText)
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
			utils.ReplyMessage(ctx, ctx.Bot(), message, "参数错误: 语义比例应为0-1的浮点数")
			return nil
		}
		queryText = strings.Join(args[:len(args)-1], " ")
	}
	artworks, err := service.SearchArtworks(ctx, &query.ArtworkSearch{
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
		common.Logger.Errorf("搜索失败: %s", err)
		utils.ReplyMessage(ctx, ctx.Bot(), message, "搜索失败, 请联系管理员检查搜索引擎设置与状态")
		return nil
	}
	if len(artworks) == 0 {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "未找到相关图片")
		return nil
	}

	if len(artworks) > 10 {
		artworks = slice.Shuffle(artworks)[:10]
	}
	handleSendResultArtworks(ctx, artworks, message, ctx.Bot())
	return nil
}

func SearchSimilarArtworks(ctx *telegohandler.Context, message telego.Message) error {
	if common.MeilisearchClient == nil {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "搜索引擎不可用")
		return nil
	}
	if message.ReplyToMessage == nil {
		helpText := `
<b>使用 /similar 命令回复一条包含图片或作品链接的消息, 将搜索与该图片相关的作品</b>

命令语法: /similar [偏移量] [限制数量]

若回复的消息中未找到支持的链接, 将尝试识别图片内容并搜索相关作品
`
		utils.ReplyMessageWithHTML(ctx, ctx.Bot(), message, helpText)
		return nil
	}
	var sourceURL string
	sourceURL = utils.FindSourceURLForMessage(message.ReplyToMessage)
	if sourceURL == "" {
		if message.ReplyToMessage.Photo == nil && message.ReplyToMessage.Document == nil {
			utils.ReplyMessage(ctx, ctx.Bot(), message, "回复的消息中未找到支持的链接")
			return nil
		}
		var err error
		var file []byte
		sourceURL, file, err = handleGetSourceURLFromPicture(ctx, message)
		if err != nil {
			common.Logger.Warnf("获取图片链接失败: %s", err)
			if file == nil || common.TaggerClient == nil {
				utils.ReplyMessage(ctx, ctx.Bot(), message, "回复的消息中未找到支持的链接或图片")
				return nil
			}
			result, err := common.TaggerClient.Predict(ctx, file)
			if err != nil || len(result.PredictedTags) == 0 {
				common.Logger.Errorf("图片识别失败: %s", err)
				utils.ReplyMessage(ctx, ctx.Bot(), message, "图片识别失败")
				return nil
			}
			queryText := strings.Join(result.PredictedTags, ",")
			artworks, err := service.SearchArtworks(ctx, &query.ArtworkSearch{
				// queryText, 0.8, 0, 10, types.R18TypeAll
				Query:  queryText,
				Hybrid: true,
				Paginate: query.Paginate{
					Offset: 0,
					Limit:  10,
				},
				R18: shared.R18TypeAll,
			})
			if err != nil || len(artworks) == 0 {
				common.Logger.Errorf("搜索失败: %s", err)
				utils.ReplyMessage(ctx, ctx.Bot(), message, "搜索失败")
				return nil
			}
			handleSendResultArtworks(ctx, artworks, message, ctx.Bot())
			return nil
		}
	}
	if sourceURL == "" {
		return nil
	}
	artwork, err := service.GetArtworkByURL(ctx, sourceURL)
	if err != nil || artwork == nil {
		common.Logger.Errorf("获取作品信息失败: %s", err)
		utils.ReplyMessage(ctx, ctx.Bot(), message, "获取作品信息失败")
		return nil
	}
	_, _, args := telegoutil.ParseCommand(message.Text)
	offset := 0
	limit := 50
	if len(args) > 0 {
		offset, err = strconv.Atoi(args[0])
		if err != nil || offset < 0 {
			utils.ReplyMessage(ctx, ctx.Bot(), message, "参数错误: 偏移量应为非负整数")
			return nil
		}
	}
	if len(args) > 1 {
		limit, err = strconv.Atoi(args[1])
		if err != nil || limit < 1 || limit > 100 {
			utils.ReplyMessage(ctx, ctx.Bot(), message, "参数错误: 限制数量应为1-10的整数")
			return nil
		}
	}
	awId, err := objectuuid.FromObjectIDHex(artwork.ID)
	if err != nil {
		common.Logger.Errorf("转换ArtworkID失败: %s", err)
		utils.ReplyMessage(ctx, ctx.Bot(), message, "内部错误")
		return nil
	}
	artworks, err := service.FindSimilarArtworks(ctx, &query.ArtworkSimilar{
		ArtworkID: awId,
		R18:       shared.R18TypeAll,
		Paginate: query.Paginate{
			Offset: offset,
			Limit:  limit,
		},
	})
	if err != nil {
		common.Logger.Errorf("搜索失败: %s", err)
		utils.ReplyMessage(ctx, ctx.Bot(), message, "搜索失败")
		return nil
	}
	if len(artworks) == 0 {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "未找到相似的作品")
		return nil
	}
	if len(artworks) > 10 {
		artworks = slice.Shuffle(artworks)[:10]
	}
	inputMedias := make([]telego.InputMedia, 0, len(artworks))
	for _, artwork := range artworks {
		picture := artwork.Pictures[0]
		var file telego.InputFile
		if picture.TelegramInfo.Data().PhotoFileID != "" {
			file = telegoutil.FileFromID(picture.TelegramInfo.Data().PhotoFileID)
		} else {
			photoURL := fmt.Sprintf("%s/?url=%s&w=2560&h=2560&we&output=jpg", config.Cfg.WSRVURL, picture.Original)
			file = telegoutil.FileFromURL(photoURL)
		}
		caption := fmt.Sprintf("<a href=\"%s\">%s</a>", artwork.SourceURL, common.EscapeHTML(artwork.Title))
		inputMedias = append(inputMedias, telegoutil.MediaPhoto(file).WithCaption(caption).WithParseMode(telego.ModeHTML))
	}
	mediaGroup := telegoutil.MediaGroup(message.Chat.ChatID(), inputMedias...).WithReplyParameters(&telego.ReplyParameters{
		MessageID: message.MessageID,
		ChatID:    message.Chat.ChatID(),
	})
	_, err = ctx.Bot().SendMediaGroup(ctx, mediaGroup)
	if err != nil {
		common.Logger.Errorf("发送图片失败: %s", err)
	}
	return nil
}

func handleGetSourceURLFromPicture(ctx *telegohandler.Context, message telego.Message) (string, []byte, error) {
	file, err := utils.GetMessagePhotoFile(ctx, ctx.Bot(), message.ReplyToMessage)
	if err != nil {
		return "", nil, err
	}
	hash, err := imgtool.GetImagePhashFromReader(bytes.NewReader(file))
	if err != nil {
		return "", file, err
	}
	pictures, err := service.GetPicturesByHashHammingDistance(ctx, hash, 10)
	if err != nil {
		return "", file, err
	}
	if len(pictures) == 0 {
		return "", file, errors.New("not found similar pictures by hash")
	}
	picture := pictures[0]
	artworkID, err := primitive.ObjectIDFromHex(picture.ArtworkID)
	if err != nil {
		return "", file, err
	}
	artwork, err := service.GetArtworkByID(ctx, artworkID)
	if err != nil {
		return "", file, err
	}
	return artwork.SourceURL, file, nil
}

func handleSendResultArtworks(ctx context.Context, artworks []*entity.Artwork, message telego.Message, bot *telego.Bot) {
	inputMedias := make([]telego.InputMedia, 0, len(artworks))
	for _, artwork := range artworks {
		picture := artwork.Pictures[0]
		var file telego.InputFile
		if picture.TelegramInfo.Data().PhotoFileID != "" {
			file = telegoutil.FileFromID(picture.TelegramInfo.Data().PhotoFileID)
		} else {
			photoURL := fmt.Sprintf("%s/?url=%s&w=2560&h=2560&we&output=jpg", config.Cfg.WSRVURL, picture.Original)
			file = telegoutil.FileFromURL(photoURL)
		}
		caption := fmt.Sprintf("<a href=\"%s\">%s</a>", artwork.SourceURL, common.EscapeHTML(artwork.Title))
		inputMedias = append(inputMedias, telegoutil.MediaPhoto(file).WithCaption(caption).WithParseMode(telego.ModeHTML))
	}
	mediaGroup := telegoutil.MediaGroup(message.Chat.ChatID(), inputMedias...).WithReplyParameters(&telego.ReplyParameters{
		MessageID: message.MessageID,
		ChatID:    message.Chat.ChatID(),
	})
	_, err := bot.SendMediaGroup(ctx, mediaGroup)
	if err != nil {
		common.Logger.Errorf("发送图片失败: %s", err)
	}
}
