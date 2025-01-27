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
	"github.com/krau/ManyACG/config"

	"github.com/krau/ManyACG/service"
	"github.com/krau/ManyACG/telegram/utils"
	"github.com/krau/ManyACG/types"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func RandomPicture(ctx context.Context, bot *telego.Bot, message telego.Message) {
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
		utils.ReplyMessage(bot, message, text)
		return
	}
	if len(artwork) == 0 {
		utils.ReplyMessage(bot, message, "未找到图片")
		return
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
	photoMessage, err := bot.SendPhoto(photo)
	if err != nil {
		utils.ReplyMessage(bot, message, "发送图片失败: "+err.Error())
	}
	if photoMessage != nil {
		picture.TelegramInfo.PhotoFileID = photoMessage.Photo[len(photoMessage.Photo)-1].FileID
		if service.UpdatePictureTelegramInfo(ctx, picture, picture.TelegramInfo) != nil {
			common.Logger.Warnf("更新图片信息失败: %s", err)
		}
	}
}

func HybridSearchArtworks(ctx context.Context, bot *telego.Bot, message telego.Message) {
	if common.MeilisearchClient == nil {
		utils.ReplyMessage(bot, message, "未启用混合搜索功能")
		return
	}
	_, _, args := telegoutil.ParseCommand(message.Text)
	if len(args) == 0 {
		utils.ReplyMessage(bot, message, "使用方法: /hybrid <搜索内容> [语义比例]\n语义比例为0-1的浮点数, 应位于参数列表最后, 越大越趋向于基于语义搜索, 若不提供, 使用默认值0.8")
		return
	}
	var hybridSemanticRatio float64
	var queryText string
	hybridSemanticRatio, err := strconv.ParseFloat(args[len(args)-1], 64)
	if err != nil {
		hybridSemanticRatio = 0.8
		queryText = strings.Join(args, " ")
	} else {
		if hybridSemanticRatio < 0 || hybridSemanticRatio > 1 {
			utils.ReplyMessage(bot, message, "参数错误: 语义比例应为0-1的浮点数")
			return
		}
		queryText = strings.Join(args[:len(args)-1], " ")
	}
	artworks, err := service.HybridSearchArtworks(ctx, queryText, hybridSemanticRatio, 0, 50)
	if err != nil {
		common.Logger.Errorf("搜索失败: %s", err)
		utils.ReplyMessage(bot, message, "搜索失败, 请联系管理员检查搜索引擎设置与状态")
		return
	}
	if len(artworks) == 0 {
		utils.ReplyMessage(bot, message, "未找到相关图片")
		return
	}

	if len(artworks) > 10 {
		artworks = slice.Shuffle(artworks)[:10]
	}
	handleSendResultArtworks(artworks, message, bot)
}

func SearchSimilarArtworks(ctx context.Context, bot *telego.Bot, message telego.Message) {
	if common.MeilisearchClient == nil {
		utils.ReplyMessage(bot, message, "搜索引擎不可用")
		return
	}
	if message.ReplyToMessage == nil {
		utils.ReplyMessage(bot, message, "请回复一条包含图片或作品链接的消息")
		return
	}
	var sourceURL string
	sourceURL = utils.FindSourceURLForMessage(message.ReplyToMessage)
	if sourceURL == "" {
		if message.ReplyToMessage.Photo == nil && message.ReplyToMessage.Document == nil {
			utils.ReplyMessage(bot, message, "回复的消息中未找到支持的链接")
			return
		}
		var err error
		var file []byte
		sourceURL, file, err = handleGetSourceURLFromPicture(ctx, bot, message)
		if err != nil {
			common.Logger.Warnf("获取图片链接失败: %s", err)
			if file == nil || common.TaggerClient == nil {
				utils.ReplyMessage(bot, message, "回复的消息中未找到支持的链接或图片")
				return
			}
			result, err := common.TaggerClient.Predict(ctx, file)
			if err != nil || len(result.PredictedTags) == 0 {
				common.Logger.Errorf("图片识别失败: %s", err)
				utils.ReplyMessage(bot, message, "图片识别失败")
				return
			}
			queryText := strings.Join(result.PredictedTags, ",")
			artworks, err := service.HybridSearchArtworks(ctx, queryText, 0.8, 0, 10)
			if err != nil || len(artworks) == 0 {
				common.Logger.Errorf("搜索失败: %s", err)
				utils.ReplyMessage(bot, message, "搜索失败")
				return
			}
			handleSendResultArtworks(artworks, message, bot)
			return
		}
	}
	if sourceURL == "" {
		return
	}
	artwork, err := service.GetArtworkByURL(ctx, sourceURL)
	if err != nil || artwork == nil {
		common.Logger.Errorf("获取作品信息失败: %s", err)
		utils.ReplyMessage(bot, message, "获取作品信息失败")
		return
	}
	_, _, args := telegoutil.ParseCommand(message.Text)
	offset := 0
	limit := 50
	if len(args) > 0 {
		offset, err = strconv.Atoi(args[0])
		if err != nil || offset < 0 {
			utils.ReplyMessage(bot, message, "参数错误: 偏移量应为非负整数")
			return
		}
	}
	if len(args) > 1 {
		limit, err = strconv.Atoi(args[1])
		if err != nil || limit < 1 || limit > 100 {
			utils.ReplyMessage(bot, message, "参数错误: 限制数量应为1-10的整数")
			return
		}
	}
	artworks, err := service.SearchSimilarArtworks(ctx, artwork.ID, int64(offset), int64(limit))
	if err != nil {
		common.Logger.Errorf("搜索失败: %s", err)
		utils.ReplyMessage(bot, message, "搜索失败")
		return
	}
	if len(artworks) == 0 {
		utils.ReplyMessage(bot, message, "未找到相似的作品")
		return
	}
	if len(artworks) > 10 {
		artworks = slice.Shuffle(artworks)[:10]
	}
	inputMedias := make([]telego.InputMedia, 0, len(artworks))
	for _, artwork := range artworks {
		picture := artwork.Pictures[0]
		var file telego.InputFile
		if picture.TelegramInfo != nil && picture.TelegramInfo.PhotoFileID != "" {
			file = telegoutil.FileFromID(picture.TelegramInfo.PhotoFileID)
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
	_, err = bot.SendMediaGroup(mediaGroup)
	if err != nil {
		common.Logger.Errorf("发送图片失败: %s", err)
	}
}

func handleGetSourceURLFromPicture(ctx context.Context, bot *telego.Bot, message telego.Message) (string, []byte, error) {
	file, err := utils.GetMessagePhotoFile(bot, message.ReplyToMessage)
	if err != nil {
		return "", nil, err
	}
	hash, err := common.GetImagePhashFromReader(bytes.NewReader(file))
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

func handleSendResultArtworks(artworks []*types.Artwork, message telego.Message, bot *telego.Bot) {
	inputMedias := make([]telego.InputMedia, 0, len(artworks))
	for _, artwork := range artworks {
		picture := artwork.Pictures[0]
		var file telego.InputFile
		if picture.TelegramInfo != nil && picture.TelegramInfo.PhotoFileID != "" {
			file = telegoutil.FileFromID(picture.TelegramInfo.PhotoFileID)
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
	_, err := bot.SendMediaGroup(mediaGroup)
	if err != nil {
		common.Logger.Errorf("发送图片失败: %s", err)
	}
}
