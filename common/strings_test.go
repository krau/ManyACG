package common

import (
	"reflect"
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
		result, err := ParseStringTo2DArray(test.input, test.sep, test.sep2)
		if err != nil {
			t.Fatalf("ParseStringTo2DArray(%s, %s, %s) failed: %v", test.input, test.sep, test.sep2, err)
		}
		if !reflect.DeepEqual(result, test.expected) {
			t.Fatalf("ParseStringTo2DArray(%s, %s, %s) = %v, expected %v", test.input, test.sep, test.sep2, result, test.expected)
		}
	}
}
