package imgtool

import (
	"os"
	"testing"
)

func TestCompressImageForTelegramByVIPS(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"test1", "input.png", "output1.jpg"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := os.ReadFile(tt.input)
			if err != nil {
				t.Fatalf("failed to read input file: %v", err)
			}
			result, err := CompressImageForTelegram(data)
			if err != nil {
				t.Fatalf("CompressImageForTelegramByVIPS() error = %v", err)
			}
			err = os.WriteFile(tt.expected, result, 0644)
			if err != nil {
				t.Fatalf("failed to write output file: %v", err)
			}
		})
	}
}

func TestCompressImageByVIPS(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		output        string
		format        string
		maxEdgeLength int
	}{
		{"test1", "test.png", "test_output1.jpg", "jpeg", 2560},
		{"test2", "test.png", "test_output2.png", "png", 2560},
		{"test3", "test.png", "test_output3.webp", "webp", 2560},
		{"test3", "test.png", "test_output4.avif", "avif", 500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CompressImageByVIPS(tt.input, tt.output, tt.format, tt.maxEdgeLength)
			if err != nil {
				t.Fatalf("CompressImageByVIPS() error = %v", err)
			}
		})
	}
}
