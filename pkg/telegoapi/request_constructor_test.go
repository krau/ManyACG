package telegoapi

import (
	"bytes"
	"io"
	"mime"
	"mime/multipart"
	"testing"

	ta "github.com/mymmrac/telego/telegoapi"
)

type namedBytes struct {
	name string
	*bytes.Reader
}

func (r namedBytes) Name() string {
	return r.name
}

func TestMultipartRequestConstructor(t *testing.T) {
	data, err := (MultipartRequestConstructor{}).MultipartRequest(
		map[string]string{"chat_id": "123", "caption": "caption"},
		map[string]ta.NamedReader{
			"media": namedBytes{
				name:   "image.jpg",
				Reader: bytes.NewReader([]byte("image data")),
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := data.BodyStream.(*io.PipeReader); ok {
		t.Fatal("multipart body must not use io.PipeReader")
	}

	payload, err := io.ReadAll(data.BodyStream)
	if err != nil {
		t.Fatal(err)
	}
	mediaType, params, err := mime.ParseMediaType(data.ContentType)
	if err != nil {
		t.Fatal(err)
	}
	if mediaType != "multipart/form-data" {
		t.Fatalf("media type = %q, want multipart/form-data", mediaType)
	}

	form := multipart.NewReader(bytes.NewReader(payload), params["boundary"])
	values := make(map[string]string)
	files := make(map[string]string)
	for {
		part, err := form.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		partData, err := io.ReadAll(part)
		if err != nil {
			t.Fatal(err)
		}
		if part.FileName() == "" {
			values[part.FormName()] = string(partData)
		} else {
			files[part.FormName()] = part.FileName() + ":" + string(partData)
		}
	}
	if values["chat_id"] != "123" || values["caption"] != "caption" {
		t.Fatalf("unexpected form values: %#v", values)
	}
	if files["media"] != "image.jpg:image data" {
		t.Fatalf("unexpected form files: %#v", files)
	}
}

func TestMultipartRequestConstructorJSONRequest(t *testing.T) {
	data, err := (MultipartRequestConstructor{}).JSONRequest(struct {
		Value string `json:"value"`
	}{Value: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if data.ContentType != ta.ContentTypeJSON {
		t.Fatalf("content type = %q, want %q", data.ContentType, ta.ContentTypeJSON)
	}
	if string(data.BodyRaw) != `{"value":"test"}` {
		t.Fatalf("body = %q", string(data.BodyRaw))
	}
}
