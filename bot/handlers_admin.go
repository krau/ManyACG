package bot

import (
	"ManyACG-Bot/config"
	"ManyACG-Bot/fetcher"
	"ManyACG-Bot/service"
	"ManyACG-Bot/storage"
	"ManyACG-Bot/telegram"
	"context"
	"errors"
	"strconv"

	. "ManyACG-Bot/logger"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/mongo"
)

func setAdmin(ctx context.Context, bot *telego.Bot, message telego.Message) {
	var userID int64
	if message.ReplyToMessage != nil {
		if message.ReplyToMessage.SenderChat != nil {
			userID = message.ReplyToMessage.SenderChat.ID
		} else {
			userID = message.ReplyToMessage.From.ID
		}
	} else {
		_, _, args := telegoutil.ParseCommand(message.Text)
		if len(args) == 0 {
			bot.SendMessage(telegoutil.Message(message.Chat.ChatID(), "请回复一条消息或提供用户ID").
				WithReplyParameters(&telego.ReplyParameters{
					MessageID: message.MessageID,
				}))
			return
		}
		var err error
		userID, err = strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			bot.SendMessage(telegoutil.Message(message.Chat.ChatID(), "请不要输入奇怪的东西").
				WithReplyParameters(&telego.ReplyParameters{
					MessageID: message.MessageID,
				}))
			return
		}
	}

	isAdmin, err := service.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			err := service.CreateAdmin(ctx, userID)
			if err != nil {
				bot.SendMessage(telegoutil.Messagef(message.Chat.ChatID(), "设置管理员失败: %s", err).
					WithReplyParameters(&telego.ReplyParameters{
						MessageID: message.MessageID,
					}))
				return
			}
			bot.SendMessage(telegoutil.Messagef(message.Chat.ChatID(), "设置管理员成功: %d", userID).
				WithReplyParameters(&telego.ReplyParameters{
					MessageID: message.MessageID,
				}))
			return
		}
		bot.SendMessage(telegoutil.Messagef(message.Chat.ChatID(), "获取管理员信息失败: %s", err).
			WithReplyParameters(&telego.ReplyParameters{
				MessageID: message.MessageID,
			}))
		return
	}
	if isAdmin {
		if err := service.DeleteAdmin(ctx, userID); err != nil {
			bot.SendMessage(telegoutil.Messagef(message.Chat.ChatID(), "删除管理员失败: %s", err).
				WithReplyParameters(&telego.ReplyParameters{
					MessageID: message.MessageID,
				}))
			return
		}
		bot.SendMessage(telegoutil.Messagef(message.Chat.ChatID(), "删除管理员成功: %d", userID).
			WithReplyParameters(&telego.ReplyParameters{
				MessageID: message.MessageID,
			}))
		return
	}

}

func deletePicture(ctx context.Context, bot *telego.Bot, message telego.Message) {
	var channelMessageID int
	if message.ReplyToMessage == nil {
		_, _, args := telegoutil.ParseCommand(message.Text)
		if len(args) == 0 {
			bot.SendMessage(telegoutil.Message(message.Chat.ChatID(), "用法:\n使用 /del 回复一条频道的图片消息, 或者提供频道消息ID").
				WithReplyParameters(&telego.ReplyParameters{
					MessageID: message.MessageID,
				}))
			return
		}
		var err error
		channelMessageID, err = strconv.Atoi(args[0])
		if err != nil {
			bot.SendMessage(telegoutil.Message(message.Chat.ChatID(), "请不要输入奇怪的东西").
				WithReplyParameters(&telego.ReplyParameters{
					MessageID: message.MessageID,
				}))
			return
		}
	} else {
		originChannel, ok := telegram.CheckTargetMessageIsChannelArtworkPost(ctx, bot, message)
		if !ok {
			bot.SendMessage(telegoutil.Message(message.Chat.ChatID(), "用法:\n使用 /del 回复一条频道的图片消息, 或者提供频道消息ID").
				WithReplyParameters(&telego.ReplyParameters{
					MessageID: message.MessageID,
				}))
			return
		}
		channelMessageID = originChannel.MessageID
	}
	picture, err := service.GetPictureByMessageID(ctx, channelMessageID)
	if err != nil {
		bot.SendMessage(telegoutil.Messagef(message.Chat.ChatID(), "获取图片信息失败: %s", err).
			WithReplyParameters(&telego.ReplyParameters{
				MessageID: message.MessageID,
			}))
		return
	}
	if err := service.DeletePictureByMessageID(ctx, channelMessageID); err != nil {
		bot.SendMessage(telegoutil.Messagef(message.Chat.ChatID(), "从数据库中删除失败: %s", err).
			WithReplyParameters(&telego.ReplyParameters{
				MessageID: message.MessageID,
			}))
		return
	}
	go bot.SendMessage(telegoutil.Messagef(message.Chat.ChatID(), "删除成功: %d", channelMessageID).
		WithReplyParameters(&telego.ReplyParameters{
			MessageID: message.MessageID,
		}))
	go bot.DeleteMessage(telegoutil.Delete(telegram.ChannelChatID, channelMessageID))

	if err := storage.GetStorage().DeletePicture(picture.StorageInfo); err != nil {
		Logger.Warnf("删除图片失败: %s", err)
		bot.SendMessage(telegoutil.Messagef(message.Chat.ChatID(), "从存储中删除失败: %s", err).
			WithReplyParameters(&telego.ReplyParameters{
				MessageID: message.MessageID,
			}))
	}
}

func fetchArtwork(ctx context.Context, bot *telego.Bot, message telego.Message) {
	go fetcher.FetchOnce(ctx, config.Cfg.Fetcher.Limit)
	bot.SendMessage(telegoutil.Message(message.Chat.ChatID(), "开始拉取作品了").
		WithReplyParameters(&telego.ReplyParameters{
			MessageID: message.MessageID,
		}))
}
