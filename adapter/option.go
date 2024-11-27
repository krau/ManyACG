package adapter

import "github.com/krau/ManyACG/types"

func OnlyLoadTag() *types.AdapterOption {
	return &types.AdapterOption{
		LoadTag: true,
	}
}

func OnlyLoadArtist() *types.AdapterOption {
	return &types.AdapterOption{
		LoadArtist: true,
	}
}

func OnlyLoadPicture() *types.AdapterOption {
	return &types.AdapterOption{
		LoadPicture: true,
	}
}

func LoadAll() *types.AdapterOption {
	return &types.AdapterOption{
		LoadTag:     true,
		LoadArtist:  true,
		LoadPicture: true,
	}
}

func LoadNone() *types.AdapterOption {
	return &types.AdapterOption{}
}

func MergeOptions(opts ...*types.AdapterOption) *types.AdapterOption {
	result := &types.AdapterOption{}
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
