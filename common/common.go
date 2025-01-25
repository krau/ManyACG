package common

import (
	"os"

	"github.com/krau/ManyACG/config"
)

func Init() {
	initHttpClient()
	initLogger()
	if config.Cfg.Auth.Resend.APIKey != "" {
		initResendClient()
	}
	if searchCfg := config.Cfg.Search; searchCfg.Enable {
		switch searchCfg.Engine {
		case "meilisearch":
			initMeilisearch()
		default:
			Logger.Fatalf("Unsupported search engine: %s", searchCfg.Engine)
			os.Exit(1)
		}
	}
	if taggerCfg := config.Cfg.Tagger; taggerCfg.Enable {
		initTaggerClient()
	}
}
