package utils

import (
	"ManyACG/types"
	"testing"
	"time"
)

var artwork = &types.Artwork{
	Title:      "【※5/12まで】受注通販のお知らせ",
	SourceType: types.SourceTypePixiv,
	Description: `コミ1新作タペストリー&amp;抱き枕カバー
	旧作抱き枕カバーの受注です！
	
	🐇あめうさぎBOOTH
	https://amedamacon.booth.pm/`,
	R18:       false,
	CreatedAt: time.Now(),
	SourceURL: "https://www.pixiv.net/artworks/118629173",
	Artist: &types.Artist{
		Name:     "飴玉コン6/30サンクリ",
		Type:     types.SourceTypePixiv,
		UID:      "1992163",
		Username: "wakasa3426",
	},
	Tags: []string{
		"Plana (BlueArchive)",
		"请问您今天要来点兔子吗？",
		"BlueArchive",
		"あめうさぎ",
		"飴玉コン",
		"サンクリ",
		"コミ1",
		"コミケ",
		"点兔",
	},
	Pictures: []*types.Picture{
		{
			Index:        0,
			Thumbnail:    "https://i.pximg.net/c/240x480/img-master/img/2021/05/10/00/00/00/118629173_p0_master1200.jpg",
			Original:     "https://i.pximg.net/img-original/img/2021/05/10/00/00/00/118629173_p0.png",
			Width:        1200,
			Height:       2400,
			Hash:         "p:e92892b764699b96",
			BlurScore:    0.0,
			TelegramInfo: &types.TelegramInfo{},
			StorageInfo:  &types.StorageInfo{},
		},
	},
}

func BenchmarkGetArtworkHTMLCaption(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetArtworkHTMLCaption(artwork)
	}
}
