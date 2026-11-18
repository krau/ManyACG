package utils

import (
	"testing"

	"github.com/mymmrac/telego"
)

func TestMessageImageFileID(t *testing.T) {
	tests := []struct {
		name    string
		message *telego.Message
		want    string
	}{
		{
			name: "photo uses largest size",
			message: &telego.Message{Photo: []telego.PhotoSize{
				{FileID: "small", Width: 160, Height: 160},
				{FileID: "large", Width: 1280, Height: 1280},
			}},
			want: "large",
		},
		{
			name: "image document uses original file",
			message: &telego.Message{Document: &telego.Document{
				FileID:    "original",
				MimeType:  "image/png",
				Thumbnail: &telego.PhotoSize{FileID: "thumbnail"},
			}},
			want: "original",
		},
		{
			name:    "image document without MIME type",
			message: &telego.Message{Document: &telego.Document{FileID: "original", FileName: "picture.jpg"}},
			want:    "original",
		},
		{
			name: "image document with generic MIME type",
			message: &telego.Message{Document: &telego.Document{
				FileID:   "original",
				FileName: "picture.WEBP",
				MimeType: "application/octet-stream",
			}},
			want: "original",
		},
		{
			name: "non-image MIME type takes precedence over filename",
			message: &telego.Message{Document: &telego.Document{
				FileID:   "document",
				FileName: "document.png",
				MimeType: "application/pdf",
			}},
		},
		{
			name:    "non-image filename without MIME type",
			message: &telego.Message{Document: &telego.Document{FileID: "document", FileName: "document.pdf"}},
		},
		{
			name:    "text message",
			message: &telego.Message{Text: "not an image"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := messageImageFileID(tt.message); got != tt.want {
				t.Errorf("messageImageFileID() = %q, want %q", got, tt.want)
			}
		})
	}
}
