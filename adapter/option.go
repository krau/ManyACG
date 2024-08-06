package adapter

type AdapterOption struct {
	LoadTag     bool
	LoadArtist  bool
	LoadPicture bool
}

func (o *AdapterOption) WithLoadTag(loadTag bool) *AdapterOption {
	o.LoadTag = loadTag
	return o
}

func (o *AdapterOption) WithLoadArtist(loadArtist bool) *AdapterOption {
	o.LoadArtist = loadArtist
	return o
}

func (o *AdapterOption) WithLoadPicture(loadPicture bool) *AdapterOption {
	o.LoadPicture = loadPicture
	return o
}

func OnlyLoadTag() *AdapterOption {
	return &AdapterOption{
		LoadTag:     true,
		LoadArtist:  false,
		LoadPicture: false,
	}
}

func OnlyLoadArtist() *AdapterOption {
	return &AdapterOption{
		LoadTag:     false,
		LoadArtist:  true,
		LoadPicture: false,
	}
}

func OnlyLoadPicture() *AdapterOption {
	return &AdapterOption{
		LoadTag:     false,
		LoadArtist:  false,
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
	return &AdapterOption{
		LoadTag:     false,
		LoadArtist:  false,
		LoadPicture: false,
	}
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
	}
	return result
}
