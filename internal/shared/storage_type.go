package shared

//go:generate go-enum --values --names --nocase

// StorageType
/*
ENUM(
webdav
local
alist
telegram
)
*/
type StorageType string
