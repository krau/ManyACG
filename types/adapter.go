package types

type AdapterOption struct {
	LoadTag          bool
	LoadArtist       bool
	LoadPicture      bool
	OnlyIndexPicture bool
}

func (o *AdapterOption) WithLoadTag() *AdapterOption {
	o.LoadTag = true
	return o
}

func (o *AdapterOption) WithLoadArtist() *AdapterOption {
	o.LoadArtist = true
	return o
}

func (o *AdapterOption) WithLoadPicture() *AdapterOption {
	o.LoadPicture = true
	return o
}

func (o *AdapterOption) WithOnlyIndexPicture() *AdapterOption {
	o.OnlyIndexPicture = true
	return o
}
