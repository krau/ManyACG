package types

type Permission string

const (
	PermissionPostArtwork    Permission = "post_artwork"
	PermissionDeleteArtwork  Permission = "delete_artwork"
	PermissionFetchArtwork   Permission = "fetch_artwork"
	PermissionGetArtworkInfo Permission = "get_artwork_info"
	PermissionSearchPicture  Permission = "search_picture"
	PermissionEditArtwork    Permission = "edit_artwork"
)

var AllPermissions = []Permission{
	PermissionPostArtwork,
	PermissionDeleteArtwork,
	PermissionFetchArtwork,
	PermissionGetArtworkInfo,
	PermissionSearchPicture,
	PermissionEditArtwork,
}
