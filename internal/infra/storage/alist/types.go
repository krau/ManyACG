package alist

import "errors"

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

type loginRequset struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Token string `json:"token"`
	} `json:"data"`
}

var (
	ErrAlistLoginFailed = errors.New("failed to login to Alist")
)
