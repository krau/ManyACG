package utils

import (
	"strings"

	"github.com/mymmrac/telego/telegoutil"
)

func ParseCommandBy(text string, splitChar, quoteChar string) (string, string, []string) {
	cmd, username, payload := telegoutil.ParseCommandPayload(text)

	if payload == "" {
		return cmd, username, []string{}
	}

	var args []string
	var currentArg strings.Builder
	inQuote := false

	for _, char := range payload {
		strChar := string(char)

		if strChar == quoteChar {
			inQuote = !inQuote
			continue
		}

		if strChar == splitChar && !inQuote {
			if currentArg.Len() > 0 {
				args = append(args, currentArg.String())
				currentArg.Reset()
			}
			continue
		}

		currentArg.WriteString(strChar)
	}

	if currentArg.Len() > 0 {
		args = append(args, currentArg.String())
	}

	return cmd, username, args
}
