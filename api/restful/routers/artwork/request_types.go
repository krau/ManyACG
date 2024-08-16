package artwork

type GetLatestArtworksRequest struct {
	R18      int   `form:"r18,default=0" binding:"gte=0,lte=2" json:"r18"`
	Page     int64 `form:"page,default=1" binding:"gte=1" json:"page"`
	PageSize int64 `form:"page_size,default=20" binding:"gte=1" json:"page_size"`
}

type GetRandomArtworksRequest struct {
	R18   int `form:"r18,default=0" binding:"gte=0,lte=2" json:"r18"`
	Limit int `form:"limit,default=1" binding:"gte=1" json:"limit"`
}

type ArtworkIDRequest struct {
	ArtworkID string `form:"artwork_id" binding:"required" json:"artwork_id"`
}
