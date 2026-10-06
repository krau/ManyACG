package handlers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/krau/ManyACG/internal/interface/telegram/handlers/utils"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/samber/oops"
)

const indexRepairProgressInterval = 10 * time.Second

func artworkIndexProgressText(progress service.ArtworkIndexProgress) string {
	stage := "检查作品"
	switch progress.Phase {
	case service.ArtworkIndexLoading:
		stage = "读取待修复作品"
	case service.ArtworkIndexIndexing:
		stage = "等待索引完成"
	}
	elapsed := time.Since(progress.StartedAt).Truncate(time.Second)
	return fmt.Sprintf("索引修复正在进行：%s\n已扫描 %d 个作品，已修复 %d 个作品\n已运行 %s", stage, progress.Scanned, progress.Repaired, elapsed)
}

func FixArtworkIndex(ctx *telegohandler.Context, message telego.Message) error {
	serv, err := requireService(ctx)
	if err != nil {
		return err
	}
	if !utils.CheckPermissionInGroup(ctx, serv, message, shared.PermissionEditArtwork) {
		utils.ReplyMessage(ctx, message, "你没有编辑作品的权限")
		return nil
	}
	if progress := serv.ArtworkIndexProgress(); progress.Running {
		_, err := utils.ReplyMessage(ctx, message, artworkIndexProgressText(progress))
		return err
	}
	msg, err := utils.ReplyMessage(ctx, message, "正在检查未索引的作品...")
	if err != nil {
		return oops.Wrapf(err, "failed to send message")
	}
	editProgress := func(editCtx context.Context, text string) error {
		_, err := ctx.Bot().EditMessageText(editCtx, &telego.EditMessageTextParams{
			ChatID:    msg.Chat.ChatID(),
			MessageID: msg.MessageID,
			Text:      text,
		})
		return err
	}
	monitorCtx, stopMonitor := context.WithCancel(ctx)
	monitorDone := make(chan struct{})
	go func() {
		defer close(monitorDone)
		ticker := time.NewTicker(indexRepairProgressInterval)
		defer ticker.Stop()
		for {
			select {
			case <-monitorCtx.Done():
				return
			case <-ticker.C:
				progress := serv.ArtworkIndexProgress()
				if progress.Running {
					if err := editProgress(monitorCtx, artworkIndexProgressText(progress)); err != nil && monitorCtx.Err() == nil {
						log.Errorf("failed to update index repair progress: %s", err)
					}
				}
			}
		}
	}()
	progress, repairErr := serv.FixArtworkIndex(ctx)
	stopMonitor()
	<-monitorDone

	var text string
	switch {
	case errors.Is(repairErr, service.ErrArtworkIndexRepairRunning):
		text = artworkIndexProgressText(progress)
	case repairErr != nil:
		log.Errorf("failed to fix artwork index: %s", repairErr)
		text = fmt.Sprintf("修复索引失败，已扫描 %d 个作品，已修复 %d 个作品；详情请查看日志", progress.Scanned, progress.Repaired)
	case progress.Repaired == 0:
		text = fmt.Sprintf("检查完成，已扫描 %d 个作品，没有发现需要补全的索引", progress.Scanned)
	default:
		text = fmt.Sprintf("索引修复完成，已扫描 %d 个作品，已修复 %d 个作品", progress.Scanned, progress.Repaired)
	}
	return editProgress(ctx, text)
}
