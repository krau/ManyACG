package utils

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/krau/ManyACG/internal/infra/kvstor"
	"github.com/krau/ManyACG/internal/interface/telegram/metautil"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/internal/shared/errs"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/krau/ManyACG/pkg/strutil"
	"github.com/mymmrac/telego"
)

const (
	artworkSubmissionKeyPrefix = "telegram:submission:"
	artworkSubmissionTTL       = 30 * 24 * time.Hour
)

type ArtworkSubmission struct {
	SourceURL   string    `msgpack:"source_url"`
	UserID      int64     `msgpack:"user_id"`
	DisplayName string    `msgpack:"display_name"`
	Username    string    `msgpack:"username"`
	SubmittedAt time.Time `msgpack:"submitted_at"`
}

func NewArtworkSubmission(sourceURL string, message telego.Message) *ArtworkSubmission {
	submission := &ArtworkSubmission{
		SourceURL:   sourceURL,
		SubmittedAt: time.Now(),
	}
	if message.SenderChat != nil {
		submission.DisplayName = strings.TrimSpace(message.SenderChat.Title)
		submission.Username = strings.TrimPrefix(message.SenderChat.Username, "@")
		// SenderChat 的 ID 是聊天 ID, 不能用于 tg://user mention.
		return submission
	}
	if message.From != nil {
		submission.UserID = message.From.ID
		submission.DisplayName = strings.TrimSpace(message.From.FirstName + " " + message.From.LastName)
		submission.Username = strings.TrimPrefix(message.From.Username, "@")
	}
	return submission
}

func artworkSubmissionKey(sourceURL string) string {
	return artworkSubmissionKeyPrefix + strutil.MD5Hash(sourceURL)
}

func SaveArtworkSubmission(ctx context.Context, store kvstor.KVStore, submission *ArtworkSubmission) error {
	if store == nil {
		return errors.New("telegram KV store is nil")
	}
	return store.Set(ctx, artworkSubmissionKey(submission.SourceURL), *submission, artworkSubmissionTTL)
}

func GetArtworkSubmission(ctx context.Context, store kvstor.KVStore, sourceURL string) (*ArtworkSubmission, error) {
	if store == nil {
		return nil, errs.ErrRecordNotFound
	}
	submission, err := kvstor.Get[ArtworkSubmission](ctx, store, artworkSubmissionKey(sourceURL))
	if err != nil {
		return nil, err
	}
	return &submission, nil
}

func DeleteArtworkSubmission(ctx context.Context, store kvstor.KVStore, sourceURL string) error {
	if store == nil {
		return errs.ErrRecordNotFound
	}
	return store.Delete(ctx, artworkSubmissionKey(sourceURL))
}

func GetArtworkSubmissionOrNil(ctx context.Context, store kvstor.KVStore, sourceURL string) *ArtworkSubmission {
	submission, err := GetArtworkSubmission(ctx, store, sourceURL)
	if err != nil {
		if !errors.Is(err, errs.ErrRecordNotFound) {
			log.Warnf("get artwork submission failed: %s", err)
		}
		return nil
	}
	return submission
}

// 优先使用 username 链接, 否则使用 tg mention.
func (s *ArtworkSubmission) MentionHTML() string {
	name := strings.TrimSpace(s.DisplayName)
	if name == "" {
		name = s.Username
	}
	if name == "" {
		name = "匿名用户"
	}
	name = EscapeHTML(name)
	if s.Username != "" {
		return fmt.Sprintf("<a href='https://t.me/%s'>%s</a>", EscapeHTML(strings.TrimPrefix(s.Username, "@")), name)
	}
	if s.UserID != 0 {
		return fmt.Sprintf("<a href='tg://user?id=%d'>%s</a>", s.UserID, name)
	}
	return name
}

func (s *ArtworkSubmission) Caption() string {
	return "via: " + s.MentionHTML()
}

// 存在投稿记录时在作品描述后追加投稿人信息.
func ChannelArtworkCaption(ctx context.Context, store kvstor.KVStore, artwork shared.ArtworkLike) string {
	caption := ArtworkHTMLCaption(artwork)
	if submission := GetArtworkSubmissionOrNil(ctx, store, artwork.GetSourceURL()); submission != nil {
		caption += "\n" + submission.Caption()
	}
	return caption
}

// SubmissionReviewChatIDs 按配置群组、管理员群组、管理员私聊的顺序返回审核目标.
func SubmissionReviewChatIDs(ctx context.Context, serv *service.Service, meta *metautil.MetaData) []telego.ChatID {
	if meta != nil {
		if groupID := meta.GroupChatID(); groupID.ID != 0 {
			return []telego.ChatID{groupID}
		}
	}
	if serv == nil {
		return nil
	}
	seen := make(map[int64]struct{})
	chats := make([]telego.ChatID, 0)
	appendID := func(id int64) {
		if id == 0 {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		chats = append(chats, telego.ChatID{ID: id})
	}
	groupIDs, err := serv.GetAdminGroupIDs(ctx)
	if err != nil {
		log.Warnf("get admin group IDs failed: %s", err)
	} else {
		for _, id := range groupIDs {
			appendID(id)
		}
	}
	if len(chats) > 0 {
		return chats
	}
	userIDs, err := serv.GetAdminUserIDs(ctx)
	if err != nil {
		log.Warnf("get admin user IDs failed: %s", err)
		return nil
	}
	for _, id := range userIDs {
		appendID(id)
	}
	return chats
}
