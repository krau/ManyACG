package handlers

import (
	"errors"
	"fmt"

	"github.com/krau/ManyACG/internal/interface/telegram/handlers/utils"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/internal/shared/errs"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/samber/oops"
)

const submitArtworkHelpText = `
<b>使用 /submit 命令并在参数中提供作品链接, 或使用该命令回复一条包含支持的链接的消息, 即可向管理员投稿作品</b>

命令语法: /submit [作品链接]
`

func SubmitArtwork(ctx *telegohandler.Context, message telego.Message) error {
	serv, err := requireService(ctx)
	if err != nil {
		return err
	}
	store, err := requireKVStore(ctx)
	if err != nil {
		return err
	}
	sourceURL := utils.FindSourceURLInMessage(serv, &message)
	if sourceURL == "" {
		sourceURL = utils.FindSourceURLInMessage(serv, message.ReplyToMessage)
	}
	if sourceURL == "" {
		utils.ReplyMessageWithHTML(ctx, message, submitArtworkHelpText)
		return nil
	}

	if artwork, err := serv.GetArtworkByURL(ctx, sourceURL); err == nil && artwork != nil {
		utils.ReplyMessage(ctx, message, "该作品已发布, 无需投稿")
		return nil
	} else if err != nil && !errors.Is(err, errs.ErrRecordNotFound) {
		return oops.Wrapf(err, "failed to get artwork by url")
	}
	if serv.CheckDeletedByURL(ctx, sourceURL) {
		utils.ReplyMessage(ctx, message, "该作品已被标记删除, 无法投稿")
		return nil
	}
	if utils.GetArtworkSubmissionOrNil(ctx, store, sourceURL) != nil {
		utils.ReplyMessage(ctx, message, "该作品已在审核中, 请耐心等待")
		return nil
	}

	cached, err := serv.GetOrFetchCachedArtwork(ctx, sourceURL)
	if err != nil {
		log.Errorf("failed to get or fetch cached artwork for submission: %s", err)
		utils.ReplyMessage(ctx, message, "获取作品信息失败, 请检查链接或稍后重试")
		return nil
	}
	sourceURL = cached.SourceURL
	if cached.Status != shared.ArtworkStatusCached {
		utils.ReplyMessage(ctx, message, "该作品已发布或正在发布中")
		return nil
	}
	if utils.GetArtworkSubmissionOrNil(ctx, store, sourceURL) != nil {
		utils.ReplyMessage(ctx, message, "该作品已在审核中, 请耐心等待")
		return nil
	}
	if artwork, err := serv.GetArtworkByURL(ctx, sourceURL); err == nil && artwork != nil {
		utils.ReplyMessage(ctx, message, "该作品已发布, 无需投稿")
		return nil
	}

	meta, err := requireMeta(ctx)
	if err != nil {
		return err
	}
	reviewChats := utils.SubmissionReviewChatIDs(ctx, serv, meta)
	if len(reviewChats) == 0 {
		utils.ReplyMessage(ctx, message, "暂时无法处理投稿")
		return nil
	}

	submission := utils.NewArtworkSubmission(sourceURL, message)
	if err := utils.SaveArtworkSubmission(ctx, store, submission); err != nil {
		return oops.Wrapf(err, "failed to save artwork submission")
	}

	sent := 0
	for _, chatID := range reviewChats {
		if err := utils.SendArtworkInfo(ctx, store, ctx.Bot(), meta, serv, sourceURL, chatID, utils.SendArtworkInfoOptions{
			HasPermission: true,
			AppendCaption: submission.Caption(),
		}); err != nil {
			log.Errorf("failed to send submission info to chat %d: %s", chatID.ID, err)
			continue
		}
		sent++
	}
	if sent == 0 {
		if err := utils.DeleteArtworkSubmission(ctx, store, sourceURL); err != nil {
			log.Errorf("failed to delete artwork submission: %s", err)
		}
		utils.ReplyMessage(ctx, message, "投稿失败, 请稍后重试")
		return nil
	}

	utils.ReplyMessageWithHTML(ctx, message, fmt.Sprintf(
		"投稿成功, 已提交审核\n作品: <a href='%s'>%s</a>",
		utils.EscapeHTML(sourceURL),
		utils.EscapeHTML(cached.GetTitle()),
	))
	return nil
}
