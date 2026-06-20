package main

import (
	"slices"
	"testing"
)

func TestCleanInput(t *testing.T) {
	tests := []struct {
		name          string
		testString    string
		expectedSlice []string
		expectedLen   int
	}{
		{
			name:          "Testing empty string",
			testString:    "",
			expectedSlice: []string{""},
			expectedLen:   1,
		},
		{
			name:          "Testing 'hello world'",
			testString:    "hello world",
			expectedSlice: []string{"hello", "world"},
			expectedLen:   2,
		},
		{
			name:          "Testing '    hello world    '",
			testString:    "    hello world    ",
			expectedSlice: []string{"hello", "world"},
			expectedLen:   2,
		},
		{
			name:          "testing 'round, dogs! hi'",
			testString:    "round, dogs! hi",
			expectedSlice: []string{"round,", "dogs!", "hi"},
			expectedLen:   3,
		},
	}

	for _, ctest := range tests {
		t.Run(ctest.name, func(t *testing.T) {
			gotslice, gotint := cleanInput(ctest.testString)
			// TODO: Look into using Google's CMP: github.com/google/go-cmp/cmp
			// Allegedly is better for deeply nested/complex structures and is quicker
			// than using reflect.DeepEqual(expected, received).
			if !slices.Equal(gotslice, ctest.expectedSlice) {
				t.Errorf("cleanInput(%s) = >>%v<<, %d -- want %v",
					ctest.testString, gotslice, gotint, ctest.expectedSlice)
			}
			if gotint != ctest.expectedLen {
				t.Errorf("cleanInput(%s) = %v, >>%d<< -- want %d",
					ctest.testString, gotslice, gotint, ctest.expectedLen)
			}
		})
	}
}
