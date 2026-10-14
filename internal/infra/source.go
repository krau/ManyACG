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

func initSource(store kvstor.KVStore) {
	nhentai.Init()
	pixiv.Init(store)
	twitter.Init()
	yandere.Init()
	bilibili.Init()
	danbooru.Init()
	kemono.Init()

	source.InitAll()
}

func sourceNamers(cfg runtimecfg.SourceConfig) map[shared.SourceType]source.ArtworkNamer {
	namers := make(map[shared.SourceType]source.ArtworkNamer, 7)
	if !cfg.Pixiv.Disable {
		namers[shared.SourceTypePixiv] = pixiv.Namer{}
	}
	if !cfg.Twitter.Disable {
		namers[shared.SourceTypeTwitter] = twitter.Namer{}
	}
	if !cfg.Yandere.Disable {
		namers[shared.SourceTypeYandere] = yandere.Namer{}
	}
	if !cfg.Bilibili.Disable {
		namers[shared.SourceTypeBilibili] = bilibili.Namer{}
	}
	if !cfg.Danbooru.Disable {
		namers[shared.SourceTypeDanbooru] = danbooru.Namer{}
	}
	if !cfg.Kemono.Disable {
		namers[shared.SourceTypeKemono] = kemono.Namer{}
	}
	if !cfg.Nhentai.Disable {
		namers[shared.SourceTypeNhentai] = nhentai.Namer{}
	}
	return namers
}
