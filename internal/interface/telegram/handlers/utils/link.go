package utils

import (
	"fmt"
	"strings"
)

func DeepLink(botUsername, command string, args ...string) string {
	return fmt.Sprintf("https://t.me/%s/?start=%s_%s", botUsername, command, strings.Join(args, "_"))
}
