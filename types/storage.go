package types

type StorageType string

const (
	StorageTypeWebdav StorageType = "webdav"
	StorageTypeLocal  StorageType = "local"
)

var StorageTypes []StorageType = []StorageType{
	StorageTypeWebdav,
	StorageTypeLocal,
}
