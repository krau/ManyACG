package handlers

import (
	"context"
	"crypto/sha256"
	"errors"
	"net/url"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
	"github.com/krau/ManyACG/internal/interface/rest/common"
	"github.com/krau/ManyACG/internal/shared/errs"
)

func NewInternalArtworkFetchHandler(fetcher common.ArtworkFetcher, token string, timeout time.Duration) fiber.Handler {
	expectedToken := sha256.Sum256([]byte(token))
	return func(ctx fiber.Ctx) error {
		if !internalTokenMatches(ctx, expectedToken) {
			return common.NewError(fiber.StatusUnauthorized, "invalid reader authorization")
		}
		var request struct {
			URL string `json:"url"`
		}
		if err := json.Unmarshal(ctx.Body(), &request); err != nil {
			return common.NewError(fiber.StatusBadRequest, "invalid artwork request")
		}
		u, err := url.Parse(request.URL)
		if err != nil || u == nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil {
			return common.NewError(fiber.StatusBadRequest, "url must be a supported artwork URL")
		}
		fetchCtx, cancel := context.WithTimeout(ctx.Context(), timeout)
		defer cancel()
		artwork, err := fetcher.GetArtworkInfo(fetchCtx, request.URL)
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(fetchCtx.Err(), context.DeadlineExceeded) {
			return common.NewError(fiber.StatusGatewayTimeout, "artwork fetch timed out")
		}
		if errors.Is(err, errs.ErrUnsupportedSource) {
			return common.NewError(fiber.StatusBadRequest, "url must be a supported artwork URL")
		}
		if err != nil || artwork == nil {
			return common.NewError(fiber.StatusBadGateway, "artwork fetch failed")
		}
		return ctx.JSON(artwork)
	}
}
