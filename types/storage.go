package types

type StorageType string

const (
	StorageTypeWebdav StorageType = "webdav"
	StorageTypeLocal  StorageType = "local"
	StorageTypeAlist  StorageType = "alist"
)

var StorageTypes []StorageType = []StorageType{
	StorageTypeWebdav,
	StorageTypeLocal,
	StorageTypeAlist,
}
