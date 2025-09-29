package shared

//go:generate go-enum --values --names --nocase

// Permission
/*
ENUM(
sudo
post_artwork
delete_artwork
get_artwork_info
edit_artwork
)
*/
type Permission string
