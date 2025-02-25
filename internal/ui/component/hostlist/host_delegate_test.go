package hostlist

import (
	"testing"

	"github.com/grafviktor/goto/internal/constant"
	"github.com/grafviktor/goto/internal/test"
	"github.com/stretchr/testify/require"
)

func TestBuildScreenLayout(t *testing.T) {
	layout := constant.ScreenLayoutDescription
	group := ""
	screenLayoutDelegate := NewHostDelegate(&layout, &group, &test.MockLogger{})
	require.Equal(t, 1, screenLayoutDelegate.Spacing())
	require.True(t, screenLayoutDelegate.ShowDescription)

	// Only when screen layout is compact - there is no spacing between
	// items and no description field is shown.
	layout = constant.ScreenLayoutCompact
	screenLayoutDelegate = NewHostDelegate(&layout, &group, &test.MockLogger{})
	require.Equal(t, 0, screenLayoutDelegate.Spacing())
	require.False(t, screenLayoutDelegate.ShowDescription)

	layout = constant.ScreenLayoutGroup
	screenLayoutDelegate = NewHostDelegate(&layout, &group, &test.MockLogger{})
	require.Equal(t, 1, screenLayoutDelegate.Spacing())
	require.True(t, screenLayoutDelegate.ShowDescription)
}

func Test_IsHostMovedToAnotherGroup(t *testing.T) {
	layout := constant.ScreenLayoutDescription
	group := ""
	hostDelegate := NewHostDelegate(&layout, &group, &test.MockLogger{})
	require.False(t, hostDelegate.isHostMovedToAnotherGroup(""))

	layout = constant.ScreenLayoutDescription
	hostDelegate = NewHostDelegate(&layout, nil, &test.MockLogger{})
	require.True(t, hostDelegate.isHostMovedToAnotherGroup("Group 1"))

	layout = constant.ScreenLayoutDescription
	group = "Group 1"
	hostDelegate = NewHostDelegate(&layout, &group, &test.MockLogger{})
	require.True(t, hostDelegate.isHostMovedToAnotherGroup("Group 2"))
}
