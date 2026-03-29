package hostlist

import (
	"strings"

	"charm.land/bubbles/v2/list"
	"github.com/samber/lo"
)

// hostListFilter - filters out host by host attributes,
// such as Title, Description and Hostname.
func hostListFilter(searchValue string, hostsDescriptionsList []string) []list.Rank {
	ranks := []list.Rank{}
	searchValue = strings.ToLower(searchValue)
	searchValue = strings.TrimSpace(searchValue)

	for index, t := range hostsDescriptionsList {
		hostAttributeList := strings.Split(strings.ToLower(t), "\n")
		for innerIndex, hostAttribute := range hostAttributeList {
			if strings.Contains(hostAttribute, searchValue) {
				matchedIndexes := []int{}
				// We only build matchedIndexes for Title. All titles
				// which contain searchStr, will be visually underlined.
				// However, we don't do that for other host attributes.
				if innerIndex == 0 {
					matchedIndexes = findMatchedIndexes(hostAttribute, searchValue)
				}

				ranks = append(ranks, list.Rank{
					Index:          index,
					MatchedIndexes: matchedIndexes,
				})

				// We only add a host once, even if all its attributes match the filter
				break
			}
		}
	}

	return ranks
}

// findMatchedIndexes - returns indexes of the matching letters, otherwise empty array.
// Example:
// str:    "abcdefghij"
// substr:   "cde"
// result: [2,3,4]
func findMatchedIndexes(str, substr string) []int {
	substrStartIdx := strings.Index(str, substr)
	if substrStartIdx < 0 {
		return []int{}
	}

	return lo.RepeatBy(len(substr), func(index int) int {
		return index + substrStartIdx
	})
}
