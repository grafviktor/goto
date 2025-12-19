package hostlist

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/samber/lo"
)

func hostListFilter(term string, targets []string) []list.Rank {
	// ranks := fuzzy.FindNoSort(term, targets)
	ranks := []list.Rank{}
	for m, t := range targets {
		values := strings.Split(t, "_$$$$$_")
		if len(values) != 0 {
			indexOfSubstring := strings.Index(values[0], term)
			if indexOfSubstring > -1 {
				ranks = append(ranks, list.Rank{
					Index: m,
					// MatchedIndexes contains every letter and it's index
					MatchedIndexes: lo.RepeatBy(len(term), func(index int) int { return index }), // int{0, 1}
				})
			}
		}

		if len(values) > 0 {
			indexOfSubstring := strings.Index(values[1], term)
			if indexOfSubstring > -1 {
				ranks = append(ranks, list.Rank{
					Index:          m,
					MatchedIndexes: []int{},
				})
			}
		}

		if len(values) > 1 {
			indexOfSubstring := strings.Index(values[2], term)
			if indexOfSubstring > -1 {
				ranks = append(ranks, list.Rank{
					Index:          m,
					MatchedIndexes: []int{},
				})
			}
		}
	}

	return ranks
}
