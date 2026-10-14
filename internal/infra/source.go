package infra

import (
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/kvstor"
	"github.com/krau/ManyACG/internal/infra/source"
	"github.com/krau/ManyACG/internal/infra/source/impls/bilibili"
	"github.com/krau/ManyACG/internal/infra/source/impls/danbooru"
	"github.com/krau/ManyACG/internal/infra/source/impls/kemono"
	"github.com/krau/ManyACG/internal/infra/source/impls/nhentai"
	"github.com/krau/ManyACG/internal/infra/source/impls/pixiv"
	"github.com/krau/ManyACG/internal/infra/source/impls/twitter"
	"github.com/krau/ManyACG/internal/infra/source/impls/yandere"
	"github.com/krau/ManyACG/internal/shared"
)

func initSource(cfg runtimecfg.Config, store kvstor.KVStore) {
	if !cfg.Source.Nhentai.Disable {
		source.RegisterNamer(shared.SourceTypeNhentai, nhentai.Namer{})
	}
	if !cfg.Source.Pixiv.Disable {
		source.RegisterNamer(shared.SourceTypePixiv, pixiv.Namer{})
	}
	if !cfg.Source.Twitter.Disable {
		source.RegisterNamer(shared.SourceTypeTwitter, twitter.Namer{})
	}
	if !cfg.Source.Yandere.Disable {
		source.RegisterNamer(shared.SourceTypeYandere, yandere.Namer{})
	}
	if !cfg.Source.Bilibili.Disable {
		source.RegisterNamer(shared.SourceTypeBilibili, bilibili.Namer{})
	}
	if !cfg.Source.Danbooru.Disable {
		source.RegisterNamer(shared.SourceTypeDanbooru, danbooru.Namer{})
	}
	if !cfg.Source.Kemono.Disable {
		source.RegisterNamer(shared.SourceTypeKemono, kemono.Namer{})
	}
	if cfg.App.Mode.ReadOnly() {
		return
	}
	nhentai.Init()
	pixiv.Init(store)
	twitter.Init()
	yandere.Init()
	bilibili.Init()
	danbooru.Init()
	kemono.Init()

	source.InitAll()
}
