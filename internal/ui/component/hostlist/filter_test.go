package hostlist

import (
	"testing"

	"charm.land/bubbles/v2/list"
	"github.com/stretchr/testify/require"
)

// Test_Filter tests the hostListFilter function with various inputs.
func Test_Filter(t *testing.T) {
	testCases := []struct {
		name                  string
		searchValue           string
		hostsDescriptionsList []string
		expected              []list.Rank
	}{
		{
			name:                  "First attribute match",
			searchValue:           "host1",
			hostsDescriptionsList: []string{"host1\nlocalhost\nThis is a test description"},
			expected:              []list.Rank{{Index: 0, MatchedIndexes: []int{0, 1, 2, 3, 4}}},
		},
		{
			name:        "When parse by value or description, MatchedIndexes should always be empty",
			searchValue: "localhost",
			hostsDescriptionsList: []string{
				"host1\nlocalhost\nThis is a test description",
				"host1\n127.0.0.1\nLocalhost",
			},
			expected: []list.Rank{
				{Index: 0, MatchedIndexes: []int{}},
				{Index: 1, MatchedIndexes: []int{}},
			},
		},
		{
			name:                  "Must give correct indexes when match found in the middle of Title",
			searchValue:           "long",
			hostsDescriptionsList: []string{"This is a long Title"},
			expected:              []list.Rank{{Index: 0, MatchedIndexes: []int{10, 11, 12, 13}}},
		},
		{
			name:                  "Dangling spaces should not be taken into account",
			searchValue:           " long ",
			hostsDescriptionsList: []string{"This is a long Title"},
			expected:              []list.Rank{{Index: 0, MatchedIndexes: []int{10, 11, 12, 13}}},
		},
		{
			name:                  "Letter casing should not matter",
			searchValue:           "LoNg",
			hostsDescriptionsList: []string{"This is a long Title"},
			expected:              []list.Rank{{Index: 0, MatchedIndexes: []int{10, 11, 12, 13}}},
		},
		{
			name:        "Must return correct index of a matching host",
			searchValue: "127",
			hostsDescriptionsList: []string{
				"host1\nlocalhost\nThis is a test description",
				"host2\n127.0.0.1\nLocalhost",
			},
			expected: []list.Rank{{Index: 1, MatchedIndexes: []int{}}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := hostListFilter(tc.searchValue, tc.hostsDescriptionsList)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func Test_FindMatchedIndexes(t *testing.T) {
	testCases := []struct {
		name     string
		str      string
		substr   string
		expected []int
	}{
		{
			name:     "Must give correct result when match found in the beginning of a word",
			str:      "abcdefg",
			substr:   "abc",
			expected: []int{0, 1, 2},
		},
		{
			name:     "Must give correct result when match found in the middle of a word",
			str:      "abcdefg",
			substr:   "bcd",
			expected: []int{1, 2, 3},
		},
		{
			name:     "Must not return a match if not all symbols from substring match",
			str:      "abcdefg",
			substr:   "abcxyz",
			expected: []int{},
		},
		{
			name:     "Must pass a silly sanity test",
			str:      "abcdefg",
			substr:   "xyz",
			expected: []int{},
		},
		{
			name:     "It is case sensitive",
			str:      "abc",
			substr:   "abC",
			expected: []int{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := findMatchedIndexes(tc.str, tc.substr)
			require.Equal(t, tc.expected, actual)
		})
	}
}
