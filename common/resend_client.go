package common

import (
	"github.com/krau/ManyACG/config"

	"github.com/resend/resend-go/v2"
)

var ResendClient *resend.Client

func initResendClient() {
	if config.Cfg.Auth.Resend.APIKey != "" {
		ResendClient = resend.NewClient(config.Cfg.Auth.Resend.APIKey)
	}
}
