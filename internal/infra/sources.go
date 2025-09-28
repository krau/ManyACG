package infra

import (
	"github.com/krau/ManyACG/internal/infra/source"
	_ "github.com/krau/ManyACG/internal/infra/source/impls/bilibili"
	_ "github.com/krau/ManyACG/internal/infra/source/impls/danbooru"
	_ "github.com/krau/ManyACG/internal/infra/source/impls/kemono"
	_ "github.com/krau/ManyACG/internal/infra/source/impls/nhentai"
	_ "github.com/krau/ManyACG/internal/infra/source/impls/pixiv"
	_ "github.com/krau/ManyACG/internal/infra/source/impls/twitter"
	_ "github.com/krau/ManyACG/internal/infra/source/impls/yandere"
)

func InitSources() {
	source.InitAll()
}
