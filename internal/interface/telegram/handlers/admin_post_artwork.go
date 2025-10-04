package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/krau/ManyACG/internal/interface/telegram/handlers/utils"
	"github.com/krau/ManyACG/internal/interface/telegram/metautil"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/krau/ManyACG/service"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/samber/oops"
)

func PostArtworkCallbackQuery(ctx *telegohandler.Context, query telego.CallbackQuery) error {
	serv := service.FromContext(ctx)
	if !utils.CheckPermissionForQuery(ctx, serv, query, shared.PermissionPostArtwork) {
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
	sourceURL, err := serv.GetStringDataByID(ctx, dataID)
	if err != nil {
		ctx.Bot().AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "获取回调数据失败 " + err.Error(),
			ShowAlert:       true,
			CacheTime:       60,
		})
		return nil
	}
	log.Infof("posting artwork: %s", sourceURL)

	cachedArtwork, err := serv.GetOrFetchCachedArtwork(ctx, sourceURL)
	if err != nil {
		ctx.Bot().AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "获取作品信息失败 " + err.Error(),
			ShowAlert:       true,
			CacheTime:       60,
		})
		return nil
	}
	if cachedArtwork.Status == shared.ArtworkStatusPosting {
		ctx.Bot().AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "该作品正在发布中",
			ShowAlert:       true,
			CacheTime:       60,
		})
		return nil
	}

	if err := serv.UpdateCachedArtworkStatusByURL(ctx, sourceURL, shared.ArtworkStatusPosting); err != nil {
		log.Errorf("更新缓存作品状态失败: %s", err)
	}
	artwork := cachedArtwork.Artwork.Data()
	ctx.Bot().EditMessageCaption(ctx, &telego.EditMessageCaptionParams{
		ChatID:      telegoutil.ID(query.Message.GetChat().ID),
		MessageID:   query.Message.GetMessageID(),
		Caption:     fmt.Sprintf("正在发布: %s", artwork.SourceURL),
		ReplyMarkup: nil,
	})
	if err := serv.CancelDeletedByURL(ctx, sourceURL); err != nil {
		log.Errorf("取消删除记录失败: %s", err)
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
	meta := metautil.FromContext(ctx)
	if meta.ChannelAvailable() {
		if err := utils.PostAndCreateArtwork(ctx, serv, artwork, query.Message.GetChat().ID, query.Message.GetMessageID()); err != nil {
			log.Errorf("failed to post and create artwork: %s", err)
			ctx.Bot().EditMessageCaption(ctx, &telego.EditMessageCaptionParams{
				ChatID:    telegoutil.ID(query.Message.GetChat().ID),
				MessageID: query.Message.GetMessageID(),
				Caption:   "发布失败: " + err.Error() + "\n" + time.Now().Format("2006-01-02 15:04:05"),
			})
			if err := serv.UpdateCachedArtworkStatusByURL(ctx, sourceURL, shared.ArtworkStatusCached); err != nil {
				// log.Warnf("更新缓存作品状态失败: %s", err)
				log.Error("failed to update cached artwork status", "err", err)
			}
		}
		return nil
	}
	return nil

	// [TODO] implement this, below is old code
	// var err error
	// createArtworkWithoutChannel := func() error {
	// 	for i, picture := range artwork.Pictures {
	// 		if len(picture.Original) > 0 {
	// 			ctx.Bot().EditMessageCaption(ctx, &telego.EditMessageCaptionParams{
	// 				ChatID:    telegoutil.ID(query.Message.GetChat().ID),
	// 				MessageID: query.Message.GetMessageID(),
	// 				ReplyMarkup: telegoutil.InlineKeyboard(
	// 					[]telego.InlineKeyboardButton{
	// 						telegoutil.InlineKeyboardButton(fmt.Sprintf("正在保存图片 %d/%d", i+1, len(artwork.Pictures))).WithCallbackData("noop"),
	// 					},
	// 				),
	// 			})
	// 		} else if i == 0 {
	// 			ctx.Bot().EditMessageCaption(ctx, &telego.EditMessageCaptionParams{
	// 				ChatID:    telegoutil.ID(query.Message.GetChat().ID),
	// 				MessageID: query.Message.GetMessageID(),
	// 				ReplyMarkup: telegoutil.InlineKeyboard(
	// 					[]telego.InlineKeyboardButton{
	// 						telegoutil.InlineKeyboardButton("正在处理存储...").WithCallbackData("noop"),
	// 					},
	// 				),
	// 			})
	// 		}
	// 		info, err := storage.SaveAll(ctx, artwork, picture)
	// 		if err != nil {
	// 			log.Errorf("保存第 %d 张图片失败: %s", i, err)
	// 			ctx.Bot().EditMessageCaption(ctx, &telego.EditMessageCaptionParams{
	// 				ChatID:    telegoutil.ID(query.Message.GetChat().ID),
	// 				MessageID: query.Message.GetMessageID(),
	// 				Caption:   "保存图片失败: " + err.Error(),
	// 			})
	// 			return err
	// 		}
	// 		artwork.Pictures[i].StorageInfo = info
	// 	}
	// 	ctx.Bot().EditMessageCaption(ctx, &telego.EditMessageCaptionParams{
	// 		ChatID:    telegoutil.ID(query.Message.GetChat().ID),
	// 		MessageID: query.Message.GetMessageID(),
	// 		ReplyMarkup: telegoutil.InlineKeyboard(
	// 			[]telego.InlineKeyboardButton{
	// 				telegoutil.InlineKeyboardButton("正在发布...").WithCallbackData("noop"),
	// 			},
	// 		),
	// 	})
	// 	artwork, err = service.CreateArtwork(ctx, artwork)
	// 	if err != nil {
	// 		log.Errorf("创建作品失败: %s", err)
	// 		ctx.Bot().EditMessageCaption(ctx, &telego.EditMessageCaptionParams{
	// 			ChatID:    telegoutil.ID(query.Message.GetChat().ID),
	// 			MessageID: query.Message.GetMessageID(),
	// 			Caption:   "创建作品失败: " + err.Error(),
	// 		})
	// 		return err
	// 	}
	// 	go func() {
	// 		for _, picture := range artwork.Pictures {
	// 			service.AddProcessPictureTask(context.TODO(), picture)
	// 		}
	// 	}()
	// 	log.Infof("Posted artwork %s", artwork.Title)
	// 	return nil
	// }
	// if err := createArtworkWithoutChannel(); err != nil {
	// 	if err := service.UpdateCachedArtworkStatusByURL(ctx, sourceURL, types.ArtworkStatusCached); err != nil {
	// 		log.Errorf("更新缓存作品状态失败: %s", err)
	// 	}
	// 	return nil
	// }

	// if err := service.UpdateCachedArtworkStatusByURL(ctx, sourceURL, types.ArtworkStatusPosted); err != nil {
	// 	log.Errorf("更新缓存作品状态失败: %s", err)
	// }
	// artwork, err = service.GetArtworkByURL(ctx, sourceURL)
	// if err != nil {
	// 	log.Errorf("获取作品信息失败: %s", err)
	// 	return nil
	// }
	// ctx.Bot().EditMessageCaption(ctx, &telego.EditMessageCaptionParams{
	// 	ChatID:      telegoutil.ID(query.Message.GetChat().ID),
	// 	MessageID:   query.Message.GetMessageID(),
	// 	Caption:     "发布成功: " + artwork.Title,
	// 	ReplyMarkup: utils.GetPostedPictureReplyMarkup(artwork, 0, ChannelChatID, BotUsername),
	// })
}

func PostArtworkCommand(ctx *telegohandler.Context, message telego.Message) error {
	serv := service.FromContext(ctx)
	if !utils.CheckPermissionInGroup(ctx, serv, message, shared.PermissionPostArtwork) {
		return oops.Errorf("user %d has no permission to post artwork", message.From.ID)
	}
	_, _, args := telegoutil.ParseCommand(message.Text)
	if len(args) == 0 && message.ReplyToMessage == nil {
		utils.ReplyMessage(ctx, message, "请提供作品链接, 或回复一条消息")
		return nil
	}
	var sourceURL string
	if message.ReplyToMessage != nil {
		sourceURL = utils.FindSourceURLInMessage(serv, message.ReplyToMessage)
		if sourceURL == "" {
			if len(args) == 0 {
				utils.ReplyMessage(ctx, message, "不支持的链接")
				return nil
			}
		}
	}
	if len(args) > 0 {
		sourceURL = serv.FindSourceURL(args[0])
	}
	if sourceURL == "" {
		utils.ReplyMessage(ctx, message, "不支持的链接")
		return nil
	}
	awEnt, _ := serv.GetArtworkByURL(ctx, sourceURL)
	if awEnt != nil {
		utils.ReplyMessage(ctx, message, "作品已存在")
		return nil
	}
	msg, err := utils.ReplyMessage(ctx, message, "正在发布...")
	if err == nil && msg != nil {
		defer ctx.Bot().DeleteMessage(ctx, telegoutil.Delete(msg.Chat.ChatID(), msg.MessageID))
	}
	cachedArtwork, err := serv.GetOrFetchCachedArtwork(ctx, sourceURL)
	if err != nil {
		log.Errorf("failed to get or fetch cached artwork: %s", err)
		utils.ReplyMessage(ctx, message, "获取作品信息失败: "+err.Error())
		return nil
	}
	if cachedArtwork.Status != shared.ArtworkStatusCached {
		utils.ReplyMessage(ctx, message, "该作品已发布或正在发布中")
		return nil
	}
	artwork := cachedArtwork.Artwork.Data()

	meta := metautil.FromContext(ctx)
	if meta.ChannelAvailable() {
		if err := utils.PostAndCreateArtwork(ctx, serv, artwork, message.Chat.ID, message.MessageID); err != nil {
			utils.ReplyMessage(ctx, message, "发布失败: "+err.Error())
			return nil
		}
		return nil
	}
	// var err error
	// for i, picture := range artwork.Pictures {
	// 	ctx.Bot().EditMessageText(ctx, &telego.EditMessageTextParams{
	// 		ChatID:    message.Chat.ChatID(),
	// 		MessageID: msg.MessageID,
	// 		Text:      fmt.Sprintf("正在保存图片 %d/%d", i+1, len(artwork.Pictures)),
	// 	})
	// 	info, err := storage.SaveAll(ctx, artwork, picture)
	// 	if err != nil {
	// 		ctx.Bot().EditMessageText(ctx, &telego.EditMessageTextParams{
	// 			ChatID:    message.Chat.ChatID(),
	// 			MessageID: msg.MessageID,
	// 			Text:      "保存图片失败: " + err.Error(),
	// 		})
	// 		return nil
	// 	}
	// 	artwork.Pictures[i].StorageInfo = info
	// }
	// ctx.Bot().EditMessageText(ctx, &telego.EditMessageTextParams{
	// 	ChatID:    message.Chat.ChatID(),
	// 	MessageID: msg.MessageID,
	// 	Text:      "图片保存完成, 正在发布...",
	// })
	// artwork, err = service.CreateArtwork(ctx, artwork)
	// if err != nil {
	// 	ctx.Bot().EditMessageText(ctx, &telego.EditMessageTextParams{
	// 		ChatID:    message.Chat.ChatID(),
	// 		MessageID: msg.MessageID,
	// 		Text:      "创建作品失败: " + err.Error(),
	// 	})
	// 	return nil
	// }
	// go func() {
	// 	for _, picture := range artwork.Pictures {
	// 		service.AddProcessPictureTask(context.TODO(), picture)
	// 	}
	// }()
	// ctx.Bot().SendMessage(ctx, telegoutil.Message(telegoutil.ID(message.Chat.ID), "发布成功: "+artwork.Title).
	// 	WithReplyParameters(&telego.ReplyParameters{
	// 		ChatID:    message.Chat.ChatID(),
	// 		MessageID: message.MessageID,
	// 	},
	// 	).WithReplyMarkup(utils.GetPostedPictureReplyMarkup(artwork, 0, ChannelChatID, BotUsername)))
	return nil
}

func ArtworkPreview(ctx *telegohandler.Context, query telego.CallbackQuery) error {
	serv := service.FromContext(ctx)
	if !utils.CheckPermissionForQuery(ctx, serv, query, shared.PermissionPostArtwork) {
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
	sourceURL, err := serv.GetStringDataByID(ctx, dataID)
	if err != nil {
		ctx.Bot().AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "获取回调数据失败 " + err.Error(),
			ShowAlert:       true,
			CacheTime:       60,
		})
		return nil
	}
	cachedArtworkEnt, err := serv.GetOrFetchCachedArtwork(ctx, sourceURL)
	if err != nil {
		ctx.Bot().AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "获取作品信息失败 " + err.Error(),
			ShowAlert:       true,
			CacheTime:       60,
		})
		return nil
	}
	if cachedArtworkEnt.Status != shared.ArtworkStatusCached {
		ctx.Bot().AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "该作品已发布或正在发布中",
			ShowAlert:       true,
			CacheTime:       60,
		})
		return nil
	}
	cachedArtwork := cachedArtworkEnt.Artwork.Data()
	var callbackMessage *telego.Message
	if query.Message.IsAccessible() {
		callbackMessage = query.Message.(*telego.Message)
	} else {
		log.Warnf("callback message is not accessible")
		return nil
	}
	meta := metautil.FromContext(ctx)
	postArtworkKeyboard := [][]telego.InlineKeyboardButton{
		{
			telegoutil.InlineKeyboardButton("发布").WithCallbackData("post_artwork " + dataID),
			telegoutil.InlineKeyboardButton("发布(!R18)").WithCallbackData("post_artwork_r18 " + dataID),
		},
		{
			telegoutil.InlineKeyboardButton("查重").WithCallbackData("search_picture " + dataID),
			telegoutil.InlineKeyboardButton("预览发布").WithURL(utils.DeepLink(meta.BotUsername, "info", dataID)),
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
		if err := serv.HideCachedArtworkPicture(ctx, cachedArtworkEnt, currentPictureIndex); err != nil {
			ctx.Bot().AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "删除失败: " + err.Error(),
				ShowAlert:       true,
				CacheTime:       60,
			})
			return nil
		}
		cachedArtworkEnt, err = serv.GetCachedArtworkByURL(ctx, sourceURL)
		if err != nil {
			ctx.Bot().AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "已删除该图片, 但获取更新信息失败: " + err.Error(),
				ShowAlert:       true,
				CacheTime:       60,
			})
			return nil
		}
		cachedArtwork = cachedArtworkEnt.Artwork.Data()
		go ctx.Bot().AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "删除成功, 稍后发布作品时将不包含该图片",
			CacheTime:       1,
		})

		previewKeyboard := []telego.InlineKeyboardButton{}

		if currentPictureIndex+1 >= len(cachedArtwork.Pictures) {
			// 如果删除的是最后一张图片, 则显示前一张
			if currentPictureIndex > 0 {
				currentPictureIndex -= 1
				currentPictureIndexStr = strconv.Itoa(currentPictureIndex)
			}
		}

		if len(cachedArtwork.Pictures) > 1 {

			deleteButton := telegoutil.InlineKeyboardButton(fmt.Sprintf("删除这张(%d)", currentPictureIndex+1)).WithCallbackData("artwork_preview " + dataID + " delete " + currentPictureIndexStr + " " + currentPictureIndexStr)
			if currentPictureIndex == 0 {
				previewKeyboard = append(previewKeyboard,
					deleteButton,
					telegoutil.InlineKeyboardButton("下一张").WithCallbackData("artwork_preview "+dataID+fmt.Sprintf(" preview %d %d", currentPictureIndex+1, currentPictureIndex)),
				)
			} else if currentPictureIndex == len(cachedArtwork.Pictures)-1 {
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
		inputFile, err := utils.GetPicturePreviewInputFile(ctx, cachedArtwork.Pictures[currentPictureIndex])
		if err != nil {
			log.Errorf("获取预览图片失败: %s", err)
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
			Media: telegoutil.MediaPhoto(inputFile).
				WithCaption(utils.ArtworkHTMLCaption(meta, cachedArtwork) + fmt.Sprintf("\n<i>当前作品有 %d 张图片</i>", len(cachedArtwork.GetPictures()))).
				WithParseMode(telego.ModeHTML),
		})
		if err != nil {
			log.Errorf("编辑预览消息失败: %s", err)
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

	inputFile, err := utils.GetPicturePreviewInputFile(ctx, cachedArtwork.Pictures[pictureIndex])
	if err != nil {
		log.Errorf("获取预览图片失败: %s", err)
		ctx.Bot().AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "获取预览图片失败: " + err.Error(),
			ShowAlert:       true,
			CacheTime:       3,
		})
		return nil
	}
	// if inputFile == nil {
	// 	go ctx.Bot().AnswerCallbackQuery(context.TODO(), &telego.AnswerCallbackQueryParams{
	// 		CallbackQueryID: query.ID,
	// 		Text:            "图片还在加载, 请稍等",
	// 		CacheTime:       3,
	// 	})
	// 	go downloadAndCompressArtwork(context.TODO(), cachedArtwork.Artwork, pictureIndex)
	// 	return nil
	// }

	previewKeyboard := []telego.InlineKeyboardButton{}
	if len(cachedArtwork.Pictures) > 1 {
		deleteButton := telegoutil.InlineKeyboardButton(fmt.Sprintf("删除这张(%d)", pictureIndex+1)).WithCallbackData("artwork_preview " + dataID + " delete " + strconv.Itoa(pictureIndex) + " " + strconv.Itoa(pictureIndex))
		if pictureIndex == 0 {
			previewKeyboard = append(previewKeyboard,
				deleteButton,
				telegoutil.InlineKeyboardButton("下一张").WithCallbackData("artwork_preview "+dataID+fmt.Sprintf(" preview %d %d", pictureIndex+1, pictureIndex)),
			)
		} else if pictureIndex == len(cachedArtwork.Pictures)-1 {
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
		Media: telegoutil.MediaPhoto(inputFile).
			WithCaption(utils.ArtworkHTMLCaption(meta, cachedArtwork) + fmt.Sprintf("\n<i>当前作品有 %d 张图片</i>", len(cachedArtwork.Pictures))).
			WithParseMode(telego.ModeHTML),
		ReplyMarkup: telegoutil.InlineKeyboard(
			postArtworkKeyboard...,
		),
	})
	if err != nil {
		log.Errorf("编辑预览消息失败: %s", err)
		return nil
	}
	cachedArtwork.Pictures[pictureIndex].TelegramInfo.PhotoFileID = msg.Photo[len(msg.Photo)-1].FileID
	if err := serv.UpdateCachedArtwork(ctx, cachedArtwork); err != nil {
		log.Errorf("更新缓存作品失败: %s", err)
	}
	return nil
}

// func downloadAndCompressArtwork(ctx context.Context, serv *service.Service, artwork entity.ArtworkLike, startIndex int) {
// 	for i, picture := range artwork.GetPictures() {
// 		if i < startIndex {
// 			continue
// 		}
// 		if picture.GetTelegramInfo().PhotoFileID != "" {
// 			continue
// 		}
// 		cachedArtwork, err := serv.GetCachedArtworkByURL(ctx, artwork.GetSourceURL())
// 		if err != nil {
// 			log.Warnf("获取缓存作品失败: %s", err)
// 			continue
// 		}
// 		if cachedArtwork.Status != shared.ArtworkStatusCached {
// 			break
// 		}
// 		tempFile, clean, err := httpclient.DownloadWithCache(ctx, picture.GetOriginal(), nil)
// 		if err != nil {
// 			log.Errorf("failed to download image: %s", err)
// 			continue
// 		}
// 		tempFile, err = imgtool.CompressImageForTelegram(tempFile)
// 		if err != nil {
// 			log.Warnf("压缩图片失败: %s", err)
// 		}
// 		cachePath := filepath.Join(config.Cfg.Storage.CacheDir, "image", common.MD5Hash(picture.Original))
// 		go common.MkCache(cachePath, tempFile, time.Duration(config.Cfg.Storage.CacheTTL)*time.Second)
// 	}
// }

// func BatchPostArtwork(ctx *telegohandler.Context, message telego.Message) error {
// 	if !CheckPermissionInGroup(ctx, message, shared.PermissionPostArtwork) {
// 		utils.ReplyMessage(ctx, ctx.Bot(), message, "你没有发布作品的权限")
// 		return nil
// 	}
// 	if message.ReplyToMessage == nil || message.ReplyToMessage.Document == nil {
// 		utils.ReplyMessage(ctx, ctx.Bot(), message, "请回复一个批量作品链接的文件")
// 		return nil
// 	}

// 	count, startIndex, sleepTime := 10, 0, 1
// 	_, _, args := telegoutil.ParseCommand(message.Text)
// 	if len(args) >= 1 {
// 		var err error
// 		count, err = strconv.Atoi(args[0])
// 		if err != nil {
// 			utils.ReplyMessage(ctx, ctx.Bot(), message, "参数错误"+err.Error())
// 			return nil
// 		}
// 	}
// 	if len(args) >= 2 {
// 		var err error
// 		startIndex, err = strconv.Atoi(args[1])
// 		if err != nil {
// 			utils.ReplyMessage(ctx, ctx.Bot(), message, "参数错误"+err.Error())
// 			return nil
// 		}
// 	}

// 	if len(args) >= 3 {
// 		var err error
// 		sleepTime, err = strconv.Atoi(args[2])
// 		if err != nil {
// 			utils.ReplyMessage(ctx, ctx.Bot(), message, "参数错误"+err.Error())
// 			return nil
// 		}
// 	}

// 	tgFile, err := ctx.Bot().GetFile(ctx, &telego.GetFileParams{
// 		FileID: message.ReplyToMessage.Document.FileID,
// 	})
// 	if err != nil {
// 		utils.ReplyMessage(ctx, ctx.Bot(), message, "获取文件失败: "+err.Error())
// 		return nil
// 	}
// 	file, err := telegoutil.DownloadFile(ctx.Bot().FileDownloadURL(tgFile.FilePath))
// 	if err != nil {
// 		utils.ReplyMessage(ctx, ctx.Bot(), message, "下载文件失败: "+err.Error())
// 		return nil
// 	}

// 	callbackMessage, err := utils.ReplyMessage(ctx, ctx.Bot(), message, fmt.Sprintf("开始发布作品...\n总数: %d\n起始索引: %d\n间隔时间: %d秒", count, startIndex, sleepTime))
// 	if err != nil {
// 		log.Warnf("回复消息失败: %s", err)
// 		callbackMessage = nil
// 	}

// 	reader := bufio.NewReader(bytes.NewReader(file))
// 	sourceURLs := make([]string, 0)

// 	for i := 0; i < startIndex; i++ {
// 		_, err := reader.ReadString('\n')
// 		if err == io.EOF {
// 			utils.ReplyMessage(ctx, ctx.Bot(), message, "起始索引超出文件行数")
// 			return nil
// 		}
// 		if err != nil {
// 			utils.ReplyMessage(ctx, ctx.Bot(), message, "读取文件失败: "+err.Error())
// 			return nil
// 		}
// 	}

// 	for i := startIndex; i < count+startIndex; i++ {
// 		text, err := reader.ReadString('\n')
// 		if err == io.EOF {
// 			utils.ReplyMessage(ctx, ctx.Bot(), message, "文件已读取完毕")
// 			break
// 		}
// 		if err != nil {
// 			utils.ReplyMessage(ctx, ctx.Bot(), message, "读取文件失败: "+err.Error())
// 			return nil
// 		}
// 		sourceURL := sources.FindSourceURL(text)
// 		if sourceURL == "" {
// 			log.Warnf("不支持的链接: %s", text)
// 			continue
// 		}
// 		sourceURLs = append(sourceURLs, sourceURL)
// 	}

// 	failed := 0
// 	for i, sourceURL := range sourceURLs {
// 		if callbackMessage != nil {
// 			if i == 0 || i%10 == 0 {
// 				ctx.Bot().EditMessageText(ctx, &telego.EditMessageTextParams{
// 					ChatID:    message.Chat.ChatID(),
// 					MessageID: callbackMessage.MessageID,
// 					Text:      fmt.Sprintf("总数: %d\n起始索引: %d\n间隔时间: %d秒\n已处理: %d\n失败: %d", count, startIndex, sleepTime, i, failed),
// 				})
// 			}
// 		}
// 		log.Infof("posting artwork: %s", sourceURL)

// 		artwork, _ := service.GetArtworkByURL(ctx, sourceURL)
// 		if artwork != nil {
// 			log.Debugf("作品已存在: %s", sourceURL)
// 			failed++
// 			continue
// 		}
// 		cachedArtwork, _ := service.GetCachedArtworkByURL(ctx, sourceURL)
// 		if cachedArtwork != nil {
// 			artwork = cachedArtwork.Artwork
// 		} else {
// 			artwork, err = sources.GetArtworkInfo(sourceURL)
// 			if err != nil {
// 				log.Errorf("获取作品信息失败: %s", err)
// 				failed++
// 				utils.ReplyMessage(ctx, ctx.Bot(), message, fmt.Sprintf("获取 %s 信息失败: %s", sourceURL, err))
// 				continue
// 			}
// 		}
// 		if err := utils.PostAndCreateArtwork(ctx, artwork, ctx.Bot(), message.Chat.ID, message.MessageID); err != nil {
// 			log.Errorf("发布失败: %s", err)
// 			failed++
// 			utils.ReplyMessage(ctx, ctx.Bot(), message, fmt.Sprintf("发布 %s 失败: %s", sourceURL, err))
// 			continue
// 		}
// 		time.Sleep(time.Duration(sleepTime) * time.Second)
// 	}
// 	if callbackMessage != nil {
// 		text := fmt.Sprintf("发布完成\n\n总数: %d\n起始索引: %d\n已处理: %d\n失败: %d", count, startIndex, count, failed)
// 		ctx.Bot().EditMessageText(ctx, &telego.EditMessageTextParams{
// 			ChatID:    message.Chat.ChatID(),
// 			MessageID: callbackMessage.MessageID,
// 			Text:      text,
// 		})
// 	}
// 	utils.ReplyMessage(ctx, ctx.Bot(), message, "批量发布作品完成")
// 	return nil
// }
