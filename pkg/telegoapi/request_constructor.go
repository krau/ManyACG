package telegoapi

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"reflect"

	ta "github.com/mymmrac/telego/telegoapi"
)

var _ ta.RequestConstructor = MultipartRequestConstructor{}

// MultipartRequestConstructor creates a streaming multipart body without an
// io.Pipe producer goroutine. File data remains in the original readers and is
// consumed only as the HTTP client reads the request body.
type MultipartRequestConstructor struct{}

func (MultipartRequestConstructor) JSONRequest(parameters any) (*ta.RequestData, error) {
	return (ta.DefaultConstructor{}).JSONRequest(parameters)
}

func (MultipartRequestConstructor) MultipartRequest(
	parameters map[string]string, filesParameters map[string]ta.NamedReader,
) (*ta.RequestData, error) {
	var headers bytes.Buffer
	writer := multipart.NewWriter(&headers)
	readers := make([]io.Reader, 0, len(filesParameters)*2+1)

	for field, value := range parameters {
		if err := writer.WriteField(field, value); err != nil {
			return nil, fmt.Errorf("write multipart field %q: %w", field, err)
		}
	}
	if headers.Len() > 0 {
		readers = append(readers, bytes.NewReader(bytes.Clone(headers.Bytes())))
		headers.Reset()
	}

	for field, file := range filesParameters {
		if isNilNamedReader(file) {
			continue
		}

		if _, err := writer.CreateFormFile(field, file.Name()); err != nil {
			return nil, fmt.Errorf("write multipart file header %q: %w", field, err)
		}
		readers = append(readers, bytes.NewReader(bytes.Clone(headers.Bytes())), file)
		headers.Reset()
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close multipart writer: %w", err)
	}
	readers = append(readers, bytes.NewReader(bytes.Clone(headers.Bytes())))

	return &ta.RequestData{
		ContentType: writer.FormDataContentType(),
		BodyStream:  io.MultiReader(readers...),
	}, nil
}

func isNilNamedReader(reader ta.NamedReader) bool {
	if reader == nil {
		return true
	}

	value := reflect.ValueOf(reader)
	switch value.Kind() {
	case reflect.Interface, reflect.Pointer, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return value.IsNil()
	default:
		return false
	}
}
