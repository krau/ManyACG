package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"os"
	"path"
	"strconv"

	"github.com/gabriel-vasile/mimetype"
	"github.com/krau/ManyACG/internal/common/httpclient"
	"github.com/krau/ManyACG/internal/interface/telegram/handlers/utils"
	"github.com/krau/ManyACG/internal/interface/telegram/metautil"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/model/query"
	"github.com/krau/ManyACG/internal/pkg/imgtool"
	"github.com/krau/ManyACG/internal/shared/errs"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/krau/ManyACG/pkg/strutil"
	"github.com/krau/ManyACG/service"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/samber/oops"
)

func GetPictureFile(ctx *telegohandler.Context, message telego.Message) error {
	var sourceURL string
	serv := service.FromContext(ctx)
	if message.ReplyToMessage != nil {
		sourceURL = utils.FindSourceURLInMessage(serv, message.ReplyToMessage)
	} else {
		sourceURL = serv.FindSourceURL(message.Text)
	}
	meta := metautil.FromContext(ctx)

	cmd, _, args := telegoutil.ParseCommand(message.Text)
	multiple := cmd == "files"
	if sourceURL == "" {
		getPictureByHash := func() *entity.Picture {
			if message.ReplyToMessage == nil {
				return nil
			}
			file, err := utils.GetMessagePhotoFile(ctx, message.ReplyToMessage)
			if err != nil {
				return nil
			}
			hash, err := imgtool.GetImagePhashFromReader(bytes.NewReader(file))
			if err != nil {
				return nil
			}
			pictures, err := serv.QueryPicturesByPhash(ctx, query.PicturesPhash{
				Input:    hash,
				Limit:    1,
				Distance: 10,
			})
			if err != nil || len(pictures) == 0 {
				return nil
			}
			return pictures[0]
		}
		picture := getPictureByHash()
		if picture == nil {
			helpText := fmt.Sprintf(`
<b>使用 /files 命令回复一条含有图片或支持的链接的消息, 或在参数中提供作品链接, 将发送作品全部原图文件</b>

命令语法: %s
`, html.EscapeString("/files [作品链接]"))
			utils.ReplyMessageWithHTML(ctx, message, helpText)
			return nil
		}
		if !multiple {
			_, err := utils.SendPictureFileByID(ctx, meta, picture.ID)
			if err != nil {
				utils.ReplyMessage(ctx, message, "文件发送失败: "+err.Error())
				return nil
			}
		} else {
			artwork, err := serv.GetArtworkByID(ctx, picture.ArtworkID)
			if err != nil {
				utils.ReplyMessage(ctx, message, "获取作品信息失败")
				return nil
			}
			getArtworkFiles(ctx, serv, meta, message, artwork)
		}
		return nil
	}

	artwork, err := serv.GetArtworkByURL(ctx, sourceURL)
	if err != nil && !errors.Is(err, errs.ErrRecordNotFound) {
		log.Errorf("获取作品信息失败: %s", err)
		utils.ReplyMessage(ctx, message, "获取作品信息失败")
		return nil
	}
	if multiple {
		if artwork == nil {
			artwork, err := serv.GetOrFetchCachedArtwork(ctx, sourceURL)
			if err != nil {
				log.Errorf("获取作品信息失败: %s", err)
				utils.ReplyMessage(ctx, message, "获取作品信息失败")
				return nil
			}
			getArtworkFiles(ctx, serv, meta, message, artwork)
			return nil
		}
		getArtworkFiles(ctx, serv, meta, message, artwork)
		return nil
	}
	if artwork == nil {
		utils.ReplyMessage(ctx, message, "未找到作品信息")
		return nil
	}
	var index int
	if len(args) != 0 {
		index, err = strconv.Atoi(args[0])
		if err != nil || index <= 0 {
			utils.ReplyMessage(ctx, message, "请输入正确的作品序号, 从 1 开始")
			return nil
		}
		index--
	}
	if index > len(artwork.Pictures) {
		utils.ReplyMessage(ctx, message, "这个作品没有这么多图片")
		return nil
	}
	picture := artwork.Pictures[index]
	_, err = utils.SendPictureFileByID(ctx, meta, picture.ID)
	if err != nil {
		if errors.Is(err, errs.ErrRecordNotFound) {
			utils.ReplyMessage(ctx, message, "这张图片未在数据库中呢")
			return nil
		}
		log.Errorf("发送文件失败: %s", err)
		utils.ReplyMessage(ctx, message, "发送文件失败, 去找管理员反馈吧~")
		return nil
	}
	return nil
}

func getArtworkFiles(ctx *telegohandler.Context, serv *service.Service,
	meta *metautil.MetaData,
	message telego.Message, artwork entity.ArtworkLike) {
	for i, picture := range artwork.GetPictures() {
		buildDocument := func() (*telego.SendDocumentParams, error) {
			var file telego.InputFile
			alreadyCached := picture.GetTelegramInfo().DocumentFileID != ""
			if alreadyCached {
				file = telegoutil.FileFromID(picture.GetTelegramInfo().DocumentFileID)
			} else if picture.GetStorageInfo().Original != nil {
				// data, err := storage.GetFile(ctx, picture.StorageInfo.Original)
				data, err := serv.Storage(picture.GetStorageInfo().Original.Type).GetFile(ctx, *picture.GetStorageInfo().Original)
				if err != nil {
					file, clean, err := httpclient.DownloadWithCache(ctx, picture.GetOriginal(), nil)
					if err != nil {
						log.Errorf("获取文件失败: %s", err)
						utils.ReplyMessage(ctx, message, fmt.Sprintf("获取第 %d 张图片失败", i+1))
						return nil, err
					}
					defer file.Close()
					defer clean()
					data, err = os.ReadFile(file.Name())
					if err != nil {
						log.Errorf("读取文件失败: %s", err)
						utils.ReplyMessage(ctx, message, fmt.Sprintf("获取第 %d 张图片失败", i+1))
						return nil, err
					}
				}
				ext, _ := strutil.GetFileExtFromURL(picture.GetOriginal())
				if ext == "" {
					mtype := mimetype.Detect(data)
					if mtype == nil {
						return nil, oops.New("failed to detect mime type")
					}
					ext = mtype.Extension()
				}
				file = telegoutil.File(telegoutil.NameReader(bytes.NewReader(data), fmt.Sprintf("%s%s", strutil.MD5Hash(picture.GetOriginal()), ext)))
			} else {
				f, clean, err := httpclient.DownloadWithCache(ctx, picture.GetOriginal(), nil)
				if err != nil {
					log.Errorf("获取文件失败: %s", err)
					utils.ReplyMessage(ctx, message, fmt.Sprintf("获取第 %d 张图片失败", i+1))
					return nil, err
				}
				defer f.Close()
				defer clean()
				data, err := os.ReadFile(f.Name())
				if err != nil {
					log.Errorf("读取文件失败: %s", err)
					utils.ReplyMessage(ctx, message, fmt.Sprintf("获取第 %d 张图片失败", i+1))
					return nil, err
				}
				file = telegoutil.File(telegoutil.NameReader(bytes.NewReader(data), path.Base(f.Name())))
			}
			document := telegoutil.Document(message.Chat.ChatID(), file).
				WithReplyParameters(&telego.ReplyParameters{
					MessageID: message.MessageID,
				}).WithCaption(artwork.GetTitle() + "_" + strconv.Itoa(i+1)).WithDisableContentTypeDetection()
			if meta.ChannelAvailable() && picture.GetTelegramInfo().MessageID != 0 {
				document.WithReplyMarkup(telegoutil.InlineKeyboard([]telego.InlineKeyboardButton{
					telegoutil.InlineKeyboardButton("详情").WithURL(meta.ChannelMessageURL(picture.GetTelegramInfo().MessageID)),
				}))
			} else {
				document.WithReplyMarkup(telegoutil.InlineKeyboard([]telego.InlineKeyboardButton{
					telegoutil.InlineKeyboardButton("详情").WithURL(artwork.GetSourceURL()),
				}))
			}
			return document, nil
		}

		document, err := buildDocument()
		if err != nil {
			break
		}
		documentMessage, err := ctx.Bot().SendDocument(ctx, document)
		if err != nil {
			log.Errorf("发送文件失败: %s", err)
			ctx.Bot().SendMessage(ctx, telegoutil.Messagef(
				message.Chat.ChatID(),
				"发送第 %d 张图片时失败",
				i+1,
			).WithReplyParameters(&telego.ReplyParameters{
				MessageID: message.MessageID,
			}))
			break
		}
		if documentMessage != nil {
			// picture.TelegramInfo.DocumentFileID = documentMessage.Document.FileID
			// if err := serv.UpdatePictureTelegramInfo(ctx, picture, picture.TelegramInfo); err != nil {
			// 	log.Warnf("更新图片信息失败: %s", err)
			// }
			// [TODO] wip
		}
		// break

	}
}
