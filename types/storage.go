package types

type StorageType string

func (s StorageType) String() string {
	return string(s)
}

const (
	StorageTypeWebdav   StorageType = "webdav"
	StorageTypeLocal    StorageType = "local"
	StorageTypeAlist    StorageType = "alist"
	StorageTypeTelegram StorageType = "telegram"
)

var StorageTypes []StorageType = []StorageType{
	StorageTypeWebdav,
	StorageTypeLocal,
	StorageTypeAlist,
	StorageTypeTelegram,
}
