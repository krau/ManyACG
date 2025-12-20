package mediatool

import (
	"encoding/base64"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/png"
	"io"

	"github.com/krau/go-thumbhash"

	_ "golang.org/x/image/webp"

	"github.com/corona10/goimagehash"
)

func GetImageThumbHash(img image.Image) (string, error) {
	tbhs := thumbhash.EncodeImage(img)
	if tbhs == nil {
		return "", fmt.Errorf("failed to encode image to thumbhash")
	}
	b64Hash := base64.StdEncoding.EncodeToString(tbhs)
	return b64Hash, nil
}

func GetImagePhashFromReader(r io.Reader) (string, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return "", err
	}
	return GetImagePhash(img)
}

func GetImagePhash(img image.Image) (string, error) {
	hash, err := goimagehash.PerceptionHash(img)
	if err != nil {
		return "", err
	}
	return hash.ToString(), nil
}
