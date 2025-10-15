package strutil

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
		result := ParseTo2DArray(test.input, test.sep, test.sep2)
		if !reflect.DeepEqual(result, test.expected) {
			t.Fatalf("ParseStringTo2DArray(%s, %s, %s) = %v, expected %v", test.input, test.sep, test.sep2, result, test.expected)
		}
	}
}

func BenchmarkParseStringTo2DArray(b *testing.B) {
	str := "1,2,3;4,5,6"
	sep := ","
	sep2 := ";"

	for b.Loop() {
		ParseTo2DArray(str, sep, sep2)
	}
}

func BenchmarkSanitizeFileName(b *testing.B) {
	fileName := strings.Repeat("test file/name\\with:illegal*chars?\"<>|%#+", 100)
	for b.Loop() {
		SanitizeFileName(fileName)
	}
}

func TestExtractTagsFromText(t *testing.T) {
	tests := []struct {
		text     string
		expected []string
	}{
		{
			text: `初音ミクHappy 16th Birthday -Dear Creators-
			✨エンドイラスト公開！✨
			https://piapro.net/miku16thbd/
			#初音ミク #miku16th`,
			expected: []string{"初音ミク", "miku16th"},
		},
		{
			text: `ひっつきむし
			#創作百合`,
			expected: []string{"創作百合"},
		},
		{
			text:     `#創作百合 #原创`,
			expected: []string{"創作百合", "原创"},
		},
		{
			text:     `プラニャ　#ブルアカ`,
			expected: []string{"ブルアカ"},
		},
		{
			text:     `原神是一款#开放世界#冒险游戏，由中国著名游戏公司#miHoYo开发。`,
			expected: []string{},
		},
	}

	for _, test := range tests {
		result := ExtractTagsFromText(test.text)
		if !reflect.DeepEqual(result, test.expected) {
			t.Fatalf("ExtractTagsFromText(%s) = %v, expected %v", test.text, result, test.expected)
		}
	}
}
