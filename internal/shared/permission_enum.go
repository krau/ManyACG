package shared

type Permission string

const (
	PermissionSudo           Permission = "sudo" // sudo can do everything
	PermissionPostArtwork    Permission = "post_artwork"
	PermissionDeleteArtwork  Permission = "delete_artwork"
	PermissionGetArtworkInfo Permission = "get_artwork_info"
	PermissionEditArtwork    Permission = "edit_artwork"
)
