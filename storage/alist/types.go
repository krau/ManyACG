package alist

type FsFormResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	// Data
}

type FsGetResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Name   string `json:"name"`
		Sign   string `json:"sign"`
		RawUrl string `json:"raw_url"`
	}
}

type FsRemoveResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	//Data
}
