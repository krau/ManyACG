package types

type Permission string

const (
	PermissionPostArtwork    Permission = "post_artwork"
	PermissionDeleteArtwork  Permission = "delete_artwork"
	PermissionDeletePicture  Permission = "delete_picture"
	PermissionFetchArtwork   Permission = "fetch_artwork"
	PermissionGetArtworkInfo Permission = "get_artwork_info"
	PermissionSearchPicture  Permission = "search_picture"
)

var AllPermissions = []Permission{
	PermissionPostArtwork,
	PermissionDeleteArtwork,
	PermissionDeletePicture,
	PermissionFetchArtwork,
	PermissionGetArtworkInfo,
	PermissionSearchPicture,
}
