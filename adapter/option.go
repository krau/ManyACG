package adapter

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

func OnlyLoadTag() *AdapterOption {
	return &AdapterOption{
		LoadTag: true,
	}
}

func OnlyLoadArtist() *AdapterOption {
	return &AdapterOption{
		LoadArtist: true,
	}
}

func OnlyLoadPicture() *AdapterOption {
	return &AdapterOption{
		LoadPicture: true,
	}
}

func LoadAll() *AdapterOption {
	return &AdapterOption{
		LoadTag:     true,
		LoadArtist:  true,
		LoadPicture: true,
	}
}

func LoadNone() *AdapterOption {
	return &AdapterOption{}
}

func MergeOptions(opts ...*AdapterOption) *AdapterOption {
	result := &AdapterOption{}
	for _, opt := range opts {
		if opt.LoadTag {
			result.LoadTag = true
		}
		if opt.LoadArtist {
			result.LoadArtist = true
		}
		if opt.LoadPicture {
			result.LoadPicture = true
		}
		if opt.OnlyIndexPicture {
			result.OnlyIndexPicture = true
		}
	}
	return result
}
