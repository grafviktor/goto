package hostlist

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grafviktor/goto/internal/constant"
	"github.com/grafviktor/goto/internal/model/host"
)

func Test_CompareTo(t *testing.T) {
	testCases := []struct {
		storageType constant.HostStorageEnum
		readonly    bool
	}{
		{
			storageType: constant.HostStorageType.YAMLFile,
			readonly:    false,
		},
		{
			storageType: constant.HostStorageType.SSHConfig,
			readonly:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(string(tc.storageType), func(t *testing.T) {
			item := ListItemHost{
				Host: host.Host{
					StorageType: tc.storageType,
				},
			}

			require.Equal(t, tc.readonly, item.ReadOnly())
		})
	}
}

func Test_ReadOnly(t *testing.T) {
	testCases := []struct {
		name     string
		id1      int
		id2      int
		title1   string
		title2   string
		expected int
	}{
		{
			name:     "Different titles, different IDs",
			id1:      1,
			title1:   "Alpha",
			id2:      2,
			title2:   "Beta",
			expected: -1,
		},
		{
			name:     "Same titles, different IDs (1 < 2)",
			id1:      1,
			title1:   "Alpha",
			id2:      2,
			title2:   "Alpha",
			expected: -1,
		},
		{
			name:     "Same titles, different IDs (2 < 1)",
			id1:      2,
			title1:   "Alpha",
			id2:      1,
			title2:   "Alpha",
			expected: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			item1 := ListItemHost{
				Host: host.Host{
					ID:    tc.id1,
					Title: tc.title1,
				},
			}
			item2 := ListItemHost{
				Host: host.Host{
					ID:    tc.id2,
					Title: tc.title2,
				},
			}

			result := item1.CompareTo(item2)
			require.Equal(t, tc.expected, result)
		})
	}
}
