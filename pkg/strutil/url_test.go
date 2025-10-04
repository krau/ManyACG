package strutil

import "testing"

func TestGetFileExtFromURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		hasError bool
	}{
		{"https://example.com/file.txt", ".txt", false},

		{"https://example.com/file.tar.gz?download=1", ".gz", false},

		{"https://example.com/报告.pdf", ".pdf", false},

		{"https://example.com/file%20name.docx", ".docx", false},

		{"https://example.com/file name.jpg", ".jpg", false},

		{"https://example.com/dir/", "", true},

		{"https://example.com/file", "", true},

		{"https://example.com", "", true},

		{"http://[::1]:namedport", "", true},
	}

	for _, tt := range tests {
		got, err := GetFileExtFromURL(tt.input)
		if (err != nil) != tt.hasError {
			t.Errorf("GetFileExtFromURL(%q) error = %v, wantError %v",
				tt.input, err, tt.hasError)
			continue
		}
		if got != tt.expected {
			t.Errorf("GetFileExtFromURL(%q) = %q, want %q",
				tt.input, got, tt.expected)
		}
	}
}
