package handlers

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/common/imgtool"
	"github.com/krau/ManyACG/config"
	"github.com/krau/ManyACG/errs"
	"github.com/krau/ManyACG/internal/model/command"
	"github.com/krau/ManyACG/internal/shared"

	"github.com/krau/ManyACG/service"
	"github.com/krau/ManyACG/sources"
	"github.com/krau/ManyACG/storage"
	"github.com/krau/ManyACG/telegram/utils"
	"github.com/krau/ManyACG/types"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
)

func PostArtworkCallbackQuery(ctx *telegohandler.Context, query telego.CallbackQuery) error {
	if !CheckPermissionForQuery(ctx, query, shared.PermissionPostArtwork) {
		ctx.Bot().AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "你没有发布作品的权限",
			ShowAlert:       true,
			CacheTime:       60,
		})
		return nil
	}
	queryDataSlice := strings.Split(query.Data, " ")
	reverseR18 := queryDataSlice[0] == "post_artwork_r18"
	dataID := queryDataSlice[1]
	sourceURL, err := service.GetCallbackDataByID(ctx, dataID)
	if err != nil {
		ctx.Bot().AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "获取回调数据失败 " + err.Error(),
			ShowAlert:       true,
			CacheTime:       60,
		})
		return nil
	}
	common.Logger.Infof("posting artwork: %s", sourceURL)

	var creation *command.ArtworkCreation
	cachedArtwork, err := service.GetCachedArtworkByURLWithCache(ctx, sourceURL)
	if err != nil {
		ctx.Bot().AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "获取作品信息失败 " + err.Error(),
			ShowAlert:       true,
			CacheTime:       60,
		})
		return nil
	}
	if cachedArtwork.Status == types.ArtworkStatusPosting {
		ctx.Bot().AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "该作品正在发布中",
			ShowAlert:       true,
			CacheTime:       60,
		})
		return nil
	}

	if err := service.UpdateCachedArtworkStatusByURL(ctx, sourceURL, types.ArtworkStatusPosting); err != nil {
		common.Logger.Errorf("更新缓存作品状态失败: %s", err)
	}
	artwork = cachedArtwork.Artwork
	go ctx.Bot().EditMessageCaption(ctx, &telego.EditMessageCaptionParams{
		ChatID:      telegoutil.ID(query.Message.GetChat().ID),
		MessageID:   query.Message.GetMessageID(),
		Caption:     fmt.Sprintf("正在发布: %s", artwork.SourceURL),
		ReplyMarkup: nil,
	})
	if err := service.DeleteDeletedByURL(ctx, sourceURL); err != nil {
		common.Logger.Errorf("取消删除记录失败: %s", err)
		ctx.Bot().EditMessageCaption(ctx, &telego.EditMessageCaptionParams{
			ChatID:    telegoutil.ID(query.Message.GetChat().ID),
			MessageID: query.Message.GetMessageID(),
			Caption:   "取消删除记录失败: " + err.Error(),
		})
		return nil
	}
	if reverseR18 {
		artwork.R18 = !artwork.R18
	}

	if IsChannelAvailable {
		if err := utils.PostAndCreateArtwork(ctx, artwork, ctx.Bot(), query.Message.GetChat().ID, query.Message.GetMessageID()); err != nil {
			common.Logger.Errorf("发布失败: %s", err)
			ctx.Bot().EditMessageCaption(ctx, &telego.EditMessageCaptionParams{
				ChatID:    telegoutil.ID(query.Message.GetChat().ID),
				MessageID: query.Message.GetMessageID(),
				Caption:   "发布失败: " + err.Error() + "\n" + time.Now().Format("2006-01-02 15:04:05"),
			})
			if err := service.UpdateCachedArtworkStatusByURL(ctx, sourceURL, types.ArtworkStatusCached); err != nil {
				common.Logger.Warnf("更新缓存作品状态失败: %s", err)
			}
			return nil
		}
	} else {
		var err error
		createArtworkWithoutChannel := func() error {
			for i, picture := range artwork.Pictures {
				if len(picture.Original) > 0 {
					ctx.Bot().EditMessageCaption(ctx, &telego.EditMessageCaptionParams{
						ChatID:    telegoutil.ID(query.Message.GetChat().ID),
						MessageID: query.Message.GetMessageID(),
						ReplyMarkup: telegoutil.InlineKeyboard(
							[]telego.InlineKeyboardButton{
								telegoutil.InlineKeyboardButton(fmt.Sprintf("正在保存图片 %d/%d", i+1, len(artwork.Pictures))).WithCallbackData("noop"),
							},
						),
					})
				} else if i == 0 {
					ctx.Bot().EditMessageCaption(ctx, &telego.EditMessageCaptionParams{
						ChatID:    telegoutil.ID(query.Message.GetChat().ID),
						MessageID: query.Message.GetMessageID(),
						ReplyMarkup: telegoutil.InlineKeyboard(
							[]telego.InlineKeyboardButton{
								telegoutil.InlineKeyboardButton("正在处理存储...").WithCallbackData("noop"),
							},
						),
					})
				}
				info, err := storage.SaveAll(ctx, artwork, picture)
				if err != nil {
					common.Logger.Errorf("保存第 %d 张图片失败: %s", i, err)
					ctx.Bot().EditMessageCaption(ctx, &telego.EditMessageCaptionParams{
						ChatID:    telegoutil.ID(query.Message.GetChat().ID),
						MessageID: query.Message.GetMessageID(),
						Caption:   "保存图片失败: " + err.Error(),
					})
					return err
				}
				artwork.Pictures[i].StorageInfo = info
			}
			ctx.Bot().EditMessageCaption(ctx, &telego.EditMessageCaptionParams{
				ChatID:    telegoutil.ID(query.Message.GetChat().ID),
				MessageID: query.Message.GetMessageID(),
				ReplyMarkup: telegoutil.InlineKeyboard(
					[]telego.InlineKeyboardButton{
						telegoutil.InlineKeyboardButton("正在发布...").WithCallbackData("noop"),
					},
				),
			})
			artwork, err = service.CreateArtwork(ctx, artwork)
			if err != nil {
				common.Logger.Errorf("创建作品失败: %s", err)
				ctx.Bot().EditMessageCaption(ctx, &telego.EditMessageCaptionParams{
					ChatID:    telegoutil.ID(query.Message.GetChat().ID),
					MessageID: query.Message.GetMessageID(),
					Caption:   "创建作品失败: " + err.Error(),
				})
				return err
			}
			go func() {
				for _, picture := range artwork.Pictures {
					service.AddProcessPictureTask(context.TODO(), picture)
				}
			}()
			common.Logger.Infof("Posted artwork %s", artwork.Title)
			return nil
		}
		if err := createArtworkWithoutChannel(); err != nil {
			if err := service.UpdateCachedArtworkStatusByURL(ctx, sourceURL, types.ArtworkStatusCached); err != nil {
				common.Logger.Errorf("更新缓存作品状态失败: %s", err)
			}
			return nil
		}
	}

	if err := service.UpdateCachedArtworkStatusByURL(ctx, sourceURL, types.ArtworkStatusPosted); err != nil {
		common.Logger.Errorf("更新缓存作品状态失败: %s", err)
	}
	artwork, err = service.GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		common.Logger.Errorf("获取作品信息失败: %s", err)
		return nil
	}
	ctx.Bot().EditMessageCaption(ctx, &telego.EditMessageCaptionParams{
		ChatID:      telegoutil.ID(query.Message.GetChat().ID),
		MessageID:   query.Message.GetMessageID(),
		Caption:     "发布成功: " + artwork.Title,
		ReplyMarkup: utils.GetPostedPictureReplyMarkup(artwork, 0, ChannelChatID, BotUsername),
	})
	return nil
}

func PostArtworkCommand(ctx *telegohandler.Context, message telego.Message) error {
	if !CheckPermissionInGroup(ctx, message, shared.PermissionPostArtwork) {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "你没有发布作品的权限")
		return nil
	}
	_, _, args := telegoutil.ParseCommand(message.Text)
	if len(args) == 0 && message.ReplyToMessage == nil {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "请提供作品链接, 或回复一条消息")
		return nil
	}
	var sourceURL string
	if message.ReplyToMessage != nil {
		sourceURL = utils.FindSourceURLForMessage(message.ReplyToMessage)
		if sourceURL == "" {
			if len(args) == 0 {
				utils.ReplyMessage(ctx, ctx.Bot(), message, "不支持的链接")
				return nil
			}
		}
	}
	if len(args) > 0 {
		sourceURL = sources.FindSourceURL(args[0])
	}
	if sourceURL == "" {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "不支持的链接")
		return nil
	}
	artwork, _ := service.GetArtworkByURL(ctx, sourceURL)
	if artwork != nil {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "作品已存在")
		return nil
	}
	msg, err := utils.ReplyMessage(ctx, ctx.Bot(), message, "正在发布...")
	if err == nil && msg != nil {
		defer ctx.Bot().DeleteMessage(ctx, telegoutil.Delete(msg.Chat.ChatID(), msg.MessageID))
	}
	cachedArtwork, err := service.GetCachedArtworkByURLWithCache(ctx, sourceURL)
	if err != nil {
		common.Logger.Errorf("获取作品信息失败: %s", err)
		utils.ReplyMessage(ctx, ctx.Bot(), message, "获取作品信息失败: "+err.Error())
		return nil
	}
	if cachedArtwork.Status != types.ArtworkStatusCached {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "该作品已发布或正在发布中")
		return nil
	}
	artwork = cachedArtwork.Artwork

	if IsChannelAvailable {
		if err := utils.PostAndCreateArtwork(ctx, artwork, ctx.Bot(), message.Chat.ID, message.MessageID); err != nil {
			utils.ReplyMessage(ctx, ctx.Bot(), message, "发布失败: "+err.Error())
			return nil
		}
	} else {
		var err error
		for i, picture := range artwork.Pictures {
			ctx.Bot().EditMessageText(ctx, &telego.EditMessageTextParams{
				ChatID:    message.Chat.ChatID(),
				MessageID: msg.MessageID,
				Text:      fmt.Sprintf("正在保存图片 %d/%d", i+1, len(artwork.Pictures)),
			})
			info, err := storage.SaveAll(ctx, artwork, picture)
			if err != nil {
				ctx.Bot().EditMessageText(ctx, &telego.EditMessageTextParams{
					ChatID:    message.Chat.ChatID(),
					MessageID: msg.MessageID,
					Text:      "保存图片失败: " + err.Error(),
				})
				return nil
			}
			artwork.Pictures[i].StorageInfo = info
		}
		ctx.Bot().EditMessageText(ctx, &telego.EditMessageTextParams{
			ChatID:    message.Chat.ChatID(),
			MessageID: msg.MessageID,
			Text:      "图片保存完成, 正在发布...",
		})
		artwork, err = service.CreateArtwork(ctx, artwork)
		if err != nil {
			ctx.Bot().EditMessageText(ctx, &telego.EditMessageTextParams{
				ChatID:    message.Chat.ChatID(),
				MessageID: msg.MessageID,
				Text:      "创建作品失败: " + err.Error(),
			})
			return nil
		}
		go func() {
			for _, picture := range artwork.Pictures {
				service.AddProcessPictureTask(context.TODO(), picture)
			}
		}()
	}

	ctx.Bot().SendMessage(ctx, telegoutil.Message(telegoutil.ID(message.Chat.ID), "发布成功: "+artwork.Title).
		WithReplyParameters(&telego.ReplyParameters{
			ChatID:    message.Chat.ChatID(),
			MessageID: message.MessageID,
		},
		).WithReplyMarkup(utils.GetPostedPictureReplyMarkup(artwork, 0, ChannelChatID, BotUsername)))
	return nil
}

func ArtworkPreview(ctx *telegohandler.Context, query telego.CallbackQuery) error {
	if !CheckPermissionForQuery(ctx, query, shared.PermissionPostArtwork) {
		ctx.Bot().AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "你没有发布作品的权限",
			ShowAlert:       true,
			CacheTime:       60,
		})
		return nil
	}
	queryDataSlice := strings.Split(query.Data, " ")
	dataID := queryDataSlice[1]
	sourceURL, err := service.GetCallbackDataByID(ctx, dataID)
	if err != nil {
		ctx.Bot().AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "获取回调数据失败 " + err.Error(),
			ShowAlert:       true,
			CacheTime:       60,
		})
		return nil
	}
	cachedArtwork, err := service.GetCachedArtworkByURLWithCache(ctx, sourceURL)
	if err != nil {
		ctx.Bot().AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "获取作品信息失败 " + err.Error(),
			ShowAlert:       true,
			CacheTime:       60,
		})
		return nil
	}
	if cachedArtwork.Status != types.ArtworkStatusCached {
		ctx.Bot().AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "该作品已发布或正在发布中",
			ShowAlert:       true,
			CacheTime:       60,
		})
		return nil
	}
	var callbackMessage *telego.Message
	if query.Message.IsAccessible() {
		callbackMessage = query.Message.(*telego.Message)
	} else {
		common.Logger.Warnf("callback message is not accessible")
		return nil
	}

	postArtworkKeyboard := [][]telego.InlineKeyboardButton{
		{
			telegoutil.InlineKeyboardButton("发布").WithCallbackData("post_artwork " + dataID),
			telegoutil.InlineKeyboardButton("发布(!R18)").WithCallbackData("post_artwork_r18 " + dataID),
		},
		{
			telegoutil.InlineKeyboardButton("查重").WithCallbackData("search_picture " + dataID),
			telegoutil.InlineKeyboardButton("预览发布").WithURL(utils.GetDeepLink(BotUsername, "info", dataID)),
		},
	}

	currentPictureIndexStr := queryDataSlice[4]
	currentPictureIndex, err := strconv.Atoi(currentPictureIndexStr)
	if err != nil {
		ctx.Bot().AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "解析回调数据错误: " + err.Error(),
			ShowAlert:       true,
			CacheTime:       60,
		})
		return nil
	}
	opera := queryDataSlice[2]
	if opera == "delete" {
		if err := service.DeleteCachedArtworkPicture(ctx, cachedArtwork, currentPictureIndex); err != nil {
			ctx.Bot().AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "删除失败: " + err.Error(),
				ShowAlert:       true,
				CacheTime:       60,
			})
			return nil
		}
		cachedArtwork, err = service.GetCachedArtworkByURL(ctx, sourceURL)
		if err != nil {
			ctx.Bot().AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "已删除该图片, 但获取更新信息失败: " + err.Error(),
				ShowAlert:       true,
				CacheTime:       60,
			})
			return nil
		}
		go ctx.Bot().AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "删除成功, 稍后发布作品时将不包含该图片",
			CacheTime:       1,
		})

		previewKeyboard := []telego.InlineKeyboardButton{}

		if currentPictureIndex+1 >= len(cachedArtwork.Artwork.Pictures) {
			// 如果删除的是最后一张图片, 则显示前一张
			if currentPictureIndex > 0 {
				currentPictureIndex -= 1
				currentPictureIndexStr = strconv.Itoa(currentPictureIndex)
			}
		}

		if len(cachedArtwork.Artwork.Pictures) > 1 {

			deleteButton := telegoutil.InlineKeyboardButton(fmt.Sprintf("删除这张(%d)", currentPictureIndex+1)).WithCallbackData("artwork_preview " + dataID + " delete " + currentPictureIndexStr + " " + currentPictureIndexStr)
			if currentPictureIndex == 0 {
				previewKeyboard = append(previewKeyboard,
					deleteButton,
					telegoutil.InlineKeyboardButton("下一张").WithCallbackData("artwork_preview "+dataID+fmt.Sprintf(" preview %d %d", currentPictureIndex+1, currentPictureIndex)),
				)
			} else if currentPictureIndex == len(cachedArtwork.Artwork.Pictures)-1 {
				previewKeyboard = append(previewKeyboard,
					telegoutil.InlineKeyboardButton("上一张").WithCallbackData("artwork_preview "+dataID+fmt.Sprintf(" preview %d %d", currentPictureIndex-1, currentPictureIndex)),
					deleteButton,
				)
			} else {
				previewKeyboard = append(previewKeyboard,
					telegoutil.InlineKeyboardButton("上一张").WithCallbackData("artwork_preview "+dataID+fmt.Sprintf(" preview %d %d", currentPictureIndex-1, currentPictureIndex)),
					deleteButton,
					telegoutil.InlineKeyboardButton("下一张").WithCallbackData("artwork_preview "+dataID+fmt.Sprintf(" preview %d %d", currentPictureIndex+1, currentPictureIndex)),
				)
			}
		}
		inputFile, err := utils.GetPicturePreviewInputFile(ctx, cachedArtwork.Artwork.Pictures[currentPictureIndex])
		if err != nil {
			common.Logger.Errorf("获取预览图片失败: %s", err)
			ctx.Bot().AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "获取预览图片失败: " + err.Error(),
				ShowAlert:       true,
				CacheTime:       60,
			})
			return nil
		}
		postArtworkKeyboard = append(postArtworkKeyboard, previewKeyboard)
		_, err = ctx.Bot().EditMessageMedia(ctx, &telego.EditMessageMediaParams{
			ChatID:      callbackMessage.Chat.ChatID(),
			MessageID:   callbackMessage.MessageID,
			ReplyMarkup: telegoutil.InlineKeyboard(postArtworkKeyboard...),
			Media: telegoutil.MediaPhoto(*inputFile).
				WithCaption(utils.GetArtworkHTMLCaption(cachedArtwork.Artwork) + fmt.Sprintf("\n<i>当前作品有 %d 张图片</i>", len(cachedArtwork.Artwork.Pictures))).
				WithParseMode(telego.ModeHTML),
		})
		if err != nil {
			common.Logger.Errorf("编辑预览消息失败: %s", err)
		}
		return nil
	}

	pictureIndex, err := strconv.Atoi(queryDataSlice[3])
	if err != nil {
		ctx.Bot().AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "解析回调数据错误: " + err.Error(),
			ShowAlert:       true,
			CacheTime:       60,
		})
		return nil
	}

	inputFile, err := utils.GetPicturePreviewInputFile(ctx, cachedArtwork.Artwork.Pictures[pictureIndex])
	if err != nil && !errors.Is(err, errs.ErrNoAvailableFile) {
		common.Logger.Errorf("获取预览图片失败: %s", err)
		ctx.Bot().AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "获取预览图片失败: " + err.Error(),
			ShowAlert:       true,
			CacheTime:       3,
		})
		return nil
	}
	if inputFile == nil {
		go ctx.Bot().AnswerCallbackQuery(context.TODO(), &telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "图片还在加载, 请稍等",
			CacheTime:       3,
		})
		go downloadAndCompressArtwork(context.TODO(), cachedArtwork.Artwork, pictureIndex)
		return nil
	}

	previewKeyboard := []telego.InlineKeyboardButton{}
	if len(cachedArtwork.Artwork.Pictures) > 1 {
		deleteButton := telegoutil.InlineKeyboardButton(fmt.Sprintf("删除这张(%d)", pictureIndex+1)).WithCallbackData("artwork_preview " + dataID + " delete " + strconv.Itoa(pictureIndex) + " " + strconv.Itoa(pictureIndex))
		if pictureIndex == 0 {
			previewKeyboard = append(previewKeyboard,
				deleteButton,
				telegoutil.InlineKeyboardButton("下一张").WithCallbackData("artwork_preview "+dataID+fmt.Sprintf(" preview %d %d", pictureIndex+1, pictureIndex)),
			)
		} else if pictureIndex == len(cachedArtwork.Artwork.Pictures)-1 {
			previewKeyboard = append(previewKeyboard,
				telegoutil.InlineKeyboardButton("上一张").WithCallbackData("artwork_preview "+dataID+fmt.Sprintf(" preview %d %d", pictureIndex-1, pictureIndex)),
				deleteButton,
			)
		} else {
			previewKeyboard = append(previewKeyboard,
				telegoutil.InlineKeyboardButton("上一张").WithCallbackData("artwork_preview "+dataID+fmt.Sprintf(" preview %d %d", pictureIndex-1, pictureIndex)),
				deleteButton,
				telegoutil.InlineKeyboardButton("下一张").WithCallbackData("artwork_preview "+dataID+fmt.Sprintf(" preview %d %d", pictureIndex+1, pictureIndex)),
			)
		}
	}
	postArtworkKeyboard = append(postArtworkKeyboard, previewKeyboard)
	msg, err := ctx.Bot().EditMessageMedia(ctx, &telego.EditMessageMediaParams{
		ChatID:    callbackMessage.Chat.ChatID(),
		MessageID: callbackMessage.MessageID,
		Media: telegoutil.MediaPhoto(*inputFile).
			WithCaption(utils.GetArtworkHTMLCaption(cachedArtwork.Artwork) + fmt.Sprintf("\n<i>当前作品有 %d 张图片</i>", len(cachedArtwork.Artwork.Pictures))).
			WithParseMode(telego.ModeHTML),
		ReplyMarkup: telegoutil.InlineKeyboard(
			postArtworkKeyboard...,
		),
	})
	if err != nil {
		common.Logger.Errorf("编辑预览消息失败: %s", err)
		return nil
	}
	if cachedArtwork.Artwork.Pictures[pictureIndex].TelegramInfo == nil {
		cachedArtwork.Artwork.Pictures[pictureIndex].TelegramInfo = &types.TelegramInfo{}
	}
	cachedArtwork.Artwork.Pictures[pictureIndex].TelegramInfo.PhotoFileID = msg.Photo[len(msg.Photo)-1].FileID
	if err := service.UpdateCachedArtwork(ctx, cachedArtwork); err != nil {
		common.Logger.Errorf("更新缓存作品失败: %s", err)
	}
	return nil
}

func downloadAndCompressArtwork(ctx context.Context, artwork *types.Artwork, startIndex int) {
	for i, picture := range artwork.Pictures {
		if i < startIndex {
			continue
		}
		if picture.TelegramInfo != nil && picture.TelegramInfo.PhotoFileID != "" {
			continue
		}
		cachedArtwork, err := service.GetCachedArtworkByURL(ctx, artwork.SourceURL)
		if err != nil {
			common.Logger.Warnf("获取缓存作品失败: %s", err)
			continue
		}
		if cachedArtwork.Status != types.ArtworkStatusCached {
			break
		}
		tempFile, err := common.DownloadWithCache(ctx, picture.Original, nil)
		if err != nil {
			common.Logger.Warnf("下载图片失败: %s", err)
			continue
		}
		tempFile, err = imgtool.CompressImageForTelegram(tempFile)
		if err != nil {
			common.Logger.Warnf("压缩图片失败: %s", err)
		}
		cachePath := filepath.Join(config.Cfg.Storage.CacheDir, "image", common.MD5Hash(picture.Original))
		go common.MkCache(cachePath, tempFile, time.Duration(config.Cfg.Storage.CacheTTL)*time.Second)
	}
}

func BatchPostArtwork(ctx *telegohandler.Context, message telego.Message) error {
	if !CheckPermissionInGroup(ctx, message, shared.PermissionPostArtwork) {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "你没有发布作品的权限")
		return nil
	}
	if message.ReplyToMessage == nil || message.ReplyToMessage.Document == nil {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "请回复一个批量作品链接的文件")
		return nil
	}

	count, startIndex, sleepTime := 10, 0, 1
	_, _, args := telegoutil.ParseCommand(message.Text)
	if len(args) >= 1 {
		var err error
		count, err = strconv.Atoi(args[0])
		if err != nil {
			utils.ReplyMessage(ctx, ctx.Bot(), message, "参数错误"+err.Error())
			return nil
		}
	}
	if len(args) >= 2 {
		var err error
		startIndex, err = strconv.Atoi(args[1])
		if err != nil {
			utils.ReplyMessage(ctx, ctx.Bot(), message, "参数错误"+err.Error())
			return nil
		}
	}

	if len(args) >= 3 {
		var err error
		sleepTime, err = strconv.Atoi(args[2])
		if err != nil {
			utils.ReplyMessage(ctx, ctx.Bot(), message, "参数错误"+err.Error())
			return nil
		}
	}

	tgFile, err := ctx.Bot().GetFile(ctx, &telego.GetFileParams{
		FileID: message.ReplyToMessage.Document.FileID,
	})
	if err != nil {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "获取文件失败: "+err.Error())
		return nil
	}
	file, err := telegoutil.DownloadFile(ctx.Bot().FileDownloadURL(tgFile.FilePath))
	if err != nil {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "下载文件失败: "+err.Error())
		return nil
	}

	callbackMessage, err := utils.ReplyMessage(ctx, ctx.Bot(), message, fmt.Sprintf("开始发布作品...\n总数: %d\n起始索引: %d\n间隔时间: %d秒", count, startIndex, sleepTime))
	if err != nil {
		common.Logger.Warnf("回复消息失败: %s", err)
		callbackMessage = nil
	}

	reader := bufio.NewReader(bytes.NewReader(file))
	sourceURLs := make([]string, 0)

	for i := 0; i < startIndex; i++ {
		_, err := reader.ReadString('\n')
		if err == io.EOF {
			utils.ReplyMessage(ctx, ctx.Bot(), message, "起始索引超出文件行数")
			return nil
		}
		if err != nil {
			utils.ReplyMessage(ctx, ctx.Bot(), message, "读取文件失败: "+err.Error())
			return nil
		}
	}

	for i := startIndex; i < count+startIndex; i++ {
		text, err := reader.ReadString('\n')
		if err == io.EOF {
			utils.ReplyMessage(ctx, ctx.Bot(), message, "文件已读取完毕")
			break
		}
		if err != nil {
			utils.ReplyMessage(ctx, ctx.Bot(), message, "读取文件失败: "+err.Error())
			return nil
		}
		sourceURL := sources.FindSourceURL(text)
		if sourceURL == "" {
			common.Logger.Warnf("不支持的链接: %s", text)
			continue
		}
		sourceURLs = append(sourceURLs, sourceURL)
	}

	failed := 0
	for i, sourceURL := range sourceURLs {
		if callbackMessage != nil {
			if i == 0 || i%10 == 0 {
				ctx.Bot().EditMessageText(ctx, &telego.EditMessageTextParams{
					ChatID:    message.Chat.ChatID(),
					MessageID: callbackMessage.MessageID,
					Text:      fmt.Sprintf("总数: %d\n起始索引: %d\n间隔时间: %d秒\n已处理: %d\n失败: %d", count, startIndex, sleepTime, i, failed),
				})
			}
		}
		common.Logger.Infof("posting artwork: %s", sourceURL)

		artwork, _ := service.GetArtworkByURL(ctx, sourceURL)
		if artwork != nil {
			common.Logger.Debugf("作品已存在: %s", sourceURL)
			failed++
			continue
		}
		cachedArtwork, _ := service.GetCachedArtworkByURL(ctx, sourceURL)
		if cachedArtwork != nil {
			artwork = cachedArtwork.Artwork
		} else {
			artwork, err = sources.GetArtworkInfo(sourceURL)
			if err != nil {
				common.Logger.Errorf("获取作品信息失败: %s", err)
				failed++
				utils.ReplyMessage(ctx, ctx.Bot(), message, fmt.Sprintf("获取 %s 信息失败: %s", sourceURL, err))
				continue
			}
		}
		if err := utils.PostAndCreateArtwork(ctx, artwork, ctx.Bot(), message.Chat.ID, message.MessageID); err != nil {
			common.Logger.Errorf("发布失败: %s", err)
			failed++
			utils.ReplyMessage(ctx, ctx.Bot(), message, fmt.Sprintf("发布 %s 失败: %s", sourceURL, err))
			continue
		}
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}
	if callbackMessage != nil {
		text := fmt.Sprintf("发布完成\n\n总数: %d\n起始索引: %d\n已处理: %d\n失败: %d", count, startIndex, count, failed)
		ctx.Bot().EditMessageText(ctx, &telego.EditMessageTextParams{
			ChatID:    message.Chat.ChatID(),
			MessageID: callbackMessage.MessageID,
			Text:      text,
		})
	}
	utils.ReplyMessage(ctx, ctx.Bot(), message, "批量发布作品完成")
	return nil
}
