package artwork

type GetArtworkListRequest struct {
	R18      int    `form:"r18,default=0" binding:"gte=0,lte=2" json:"r18"`
	ArtistID string `form:"artist_id" binding:"omitempty" json:"artist_id"`
	Tag      string `form:"tag" binding:"omitempty" json:"tag"`
	Page     int64  `form:"page,default=1" binding:"gte=1" json:"page"`
	PageSize int64  `form:"page_size,default=20" binding:"gte=1,lte=200" json:"page_size"`
}

type GetRandomArtworksRequest struct {
	R18   int `form:"r18,default=0" binding:"gte=0,lte=2" json:"r18"`
	Limit int `form:"limit,default=1" binding:"gte=1,lte=200" json:"limit"`
}

type ArtworkIDRequest struct {
	ArtworkID string `form:"artwork_id" binding:"required" json:"artwork_id"`
}

type R18Request struct {
	R18 int `form:"r18,default=0" binding:"gte=0,lte=2" json:"r18"`
}
