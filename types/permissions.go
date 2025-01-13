package types

type Permission string

const (
	PermissionPostArtwork    Permission = "post_artwork"
	PermissionDeleteArtwork  Permission = "delete_artwork"
	PermissionGetArtworkInfo Permission = "get_artwork_info"
	PermissionEditArtwork    Permission = "edit_artwork"

	// deprecated
	PermissionFetchArtwork  Permission = "fetch_artwork"
	PermissionSearchPicture Permission = "search_picture"
)

var AllPermissions = []Permission{
	PermissionPostArtwork,
	PermissionDeleteArtwork,
	PermissionGetArtworkInfo,
	PermissionEditArtwork,

	// deprecated
	PermissionFetchArtwork,
	PermissionSearchPicture,
}
