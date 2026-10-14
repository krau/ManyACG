package handlers

import (
	"bufio"
	"context"
	"crypto/sha256"
	"errors"
	"io"
	"net/url"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
	"github.com/krau/ManyACG/internal/interface/rest/common"
)

func NewInternalArtworkDownloadHandler(downloader common.ArtworkDownloader, token string, timeout time.Duration) fiber.Handler {
	expectedToken := sha256.Sum256([]byte(token))
	return func(ctx fiber.Ctx) error {
		if !internalTokenMatches(ctx, expectedToken) {
			return common.NewError(fiber.StatusUnauthorized, "invalid reader authorization")
		}
		var request struct {
			URL string `json:"url"`
		}
		if err := json.Unmarshal(ctx.Body(), &request); err != nil {
			return common.NewError(fiber.StatusBadRequest, "invalid download request")
		}
		u, err := url.Parse(request.URL)
		if err != nil || u == nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil {
			return common.NewError(fiber.StatusBadRequest, "url must be a supported http(s) URL")
		}
		downloadCtx, cancel := context.WithTimeout(ctx.Context(), timeout)
		defer cancel()
		file, err := downloader.DownloadOriginFile(downloadCtx, request.URL)
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(downloadCtx.Err(), context.DeadlineExceeded) {
			return common.NewError(fiber.StatusGatewayTimeout, "file download timed out")
		}
		if err != nil || file == nil {
			return common.NewError(fiber.StatusBadGateway, "file download failed")
		}
		ctx.Set(fiber.HeaderContentType, fiber.MIMEOctetStream)
		return ctx.SendStreamWriter(func(w *bufio.Writer) {
			defer file.Close()
			defer w.Flush()
			_, _ = io.Copy(w, file)
		})
	}
}
