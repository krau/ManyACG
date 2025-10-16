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

// 为兼容性保留了 alist 存储类型定义, 但实现已被移除
