package handlers

import (
	"crypto/sha256"
	"crypto/subtle"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// internalTokenMatches reports whether the request carries the expected bearer
// token, comparing digests in constant time.
func internalTokenMatches(ctx fiber.Ctx, expected [sha256.Size]byte) bool {
	providedToken, isBearer := strings.CutPrefix(ctx.Get(fiber.HeaderAuthorization), "Bearer ")
	providedDigest := sha256.Sum256([]byte(providedToken))
	matches := subtle.ConstantTimeCompare(providedDigest[:], expected[:]) == 1
	return matches && isBearer
}
