package handlers

import (
	"github.com/krau/ManyACG/internal/interface/telegram/metautil"
	"github.com/krau/ManyACG/internal/service"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/samber/oops"
)

func requireService(ctx *telegohandler.Context) (*service.Service, error) {
	serv := service.FromContext(ctx)
	if serv == nil {
		return nil, oops.New("telegram handler missing service in context")
	}
	return serv, nil
}

func requireMeta(ctx *telegohandler.Context) (*metautil.MetaData, error) {
	meta := metautil.FromContext(ctx)
	if meta == nil {
		return nil, oops.New("telegram handler missing metadata in context")
	}
	return meta, nil
}

func requireSourceURL(ctx *telegohandler.Context) (string, error) {
	value := ctx.Value("source_url")
	sourceURL, ok := value.(string)
	if !ok || sourceURL == "" {
		return "", oops.New("telegram handler missing source_url in context")
	}
	return sourceURL, nil
}
