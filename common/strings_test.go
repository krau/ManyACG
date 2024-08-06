package common

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseStringTo2DArray(t *testing.T) {
	tests := []struct {
		input    string
		sep      string
		sep2     string
		expected [][]string
	}{
		{
			input:    "1,2,3;4,5,6",
			sep:      ",",
			sep2:     ";",
			expected: [][]string{{"1", "2", "3"}, {"4", "5", "6"}},
		},
		{
			input:    "1,2,3;\"4,5,6\"",
			sep:      ",",
			sep2:     ";",
			expected: [][]string{{"1", "2", "3"}, {"4,5,6"}},
		},
		{
			input:    "'1,2',3;4,5,6",
			sep:      ",",
			sep2:     ";",
			expected: [][]string{{"1,2", "3"}, {"4", "5", "6"}},
		},
		{
			input:    "1,2,'3;4',5;6,7,8",
			sep:      ",",
			sep2:     ";",
			expected: [][]string{{"1", "2", "3;4", "5"}, {"6", "7", "8"}},
		},
		{
			input:    "1,2,3",
			sep:      ",",
			sep2:     ";",
			expected: [][]string{{"1", "2", "3"}},
		},
		{
			input:    "",
			sep:      ",",
			sep2:     ";",
			expected: nil,
		},
	}

	for _, test := range tests {
		result := ParseStringTo2DArray(test.input, test.sep, test.sep2)
		if !reflect.DeepEqual(result, test.expected) {
			t.Fatalf("ParseStringTo2DArray(%s, %s, %s) = %v, expected %v", test.input, test.sep, test.sep2, result, test.expected)
		}
	}
}

func BenchmarkParseStringTo2DArray(b *testing.B) {
	str := "1,2,3;4,5,6"
	sep := ","
	sep2 := ";"

	for i := 0; i < b.N; i++ {
		ParseStringTo2DArray(str, sep, sep2)
	}
}

func BenchmarkReplaceFileNameInvalidChar(b *testing.B) {
	fileName := strings.Repeat("test file/name\\with:illegal*chars?\"<>|%#+", 100)
	for i := 0; i < b.N; i++ {
		ReplaceFileNameInvalidChar(fileName)
	}
}

func BenchmarkEscapeHTML(b *testing.B) {
	text := strings.Repeat(`This is a <test> & it should be escaped`, 100)

	for i := 0; i < b.N; i++ {
		EscapeHTML(text)
	}
}
