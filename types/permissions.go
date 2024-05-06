package types

type Permission string

const (
	PostArtwork    Permission = "post_artwork"
	DeleteArtwork  Permission = "delete_artwork"
	DeletePicture  Permission = "delete_picture"
	FetchArtwork   Permission = "fetch_artwork"
	GetArtworkInfo Permission = "get_artwork_info"
	SearchPicture  Permission = "search_picture"
)

var AllPermissions = []Permission{
	PostArtwork,
	DeleteArtwork,
	DeletePicture,
	FetchArtwork,
	GetArtworkInfo,
	SearchPicture,
}
