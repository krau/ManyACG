package handlers

import (
	"github.com/krau/ManyACG/internal/common/restclient"
	"github.com/krau/ManyACG/internal/infra/kvstor"
	"github.com/krau/ManyACG/internal/interface/telegram/handlers/utils"
	"github.com/krau/ManyACG/internal/interface/telegram/metautil"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/samber/oops"
)

type kvStoreContextKey struct{}

type restClientContextKey struct{}

func requireKVStore(ctx *telegohandler.Context) (kvstor.KVStore, error) {
	store, ok := ctx.Value(kvStoreContextKey{}).(kvstor.KVStore)
	if !ok || store == nil {
		return nil, oops.New("telegram handler missing KV store in context")
	}
	return store, nil
}

func channelArtworkCaption(ctx *telegohandler.Context, artwork shared.ArtworkLike) string {
	store, err := requireKVStore(ctx)
	if err != nil {
		return utils.ArtworkHTMLCaption(artwork)
	}
	return utils.ChannelArtworkCaption(ctx, store, artwork)
}

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

func restClient(ctx *telegohandler.Context) *restclient.Client {
	client, _ := ctx.Value(restClientContextKey{}).(*restclient.Client)
	return client
}
