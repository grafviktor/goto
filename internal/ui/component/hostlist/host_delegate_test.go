package hostlist

import (
	"bytes"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/grafviktor/goto/internal/constant"
	"github.com/grafviktor/goto/internal/model/host"
	testutils "github.com/grafviktor/goto/internal/testutils"
)

func TestBuildScreenLayout(t *testing.T) {
	layout := constant.ScreenLayoutDescription
	group := ""
	screenLayoutDelegate := NewHostDelegate(&layout, &group, &testutils.MockLogger{})
	require.Equal(t, 1, screenLayoutDelegate.Spacing())
	require.True(t, screenLayoutDelegate.ShowDescription)

	// Only when screen layout is compact - there is no spacing between
	// items and no description field is shown.
	layout = constant.ScreenLayoutCompact
	screenLayoutDelegate = NewHostDelegate(&layout, &group, &testutils.MockLogger{})
	require.Equal(t, 0, screenLayoutDelegate.Spacing())
	require.False(t, screenLayoutDelegate.ShowDescription)

	layout = constant.ScreenLayoutGroup
	screenLayoutDelegate = NewHostDelegate(&layout, &group, &testutils.MockLogger{})
	require.Equal(t, 1, screenLayoutDelegate.Spacing())
	require.True(t, screenLayoutDelegate.ShowDescription)
}

func Test_IsHostMovedToAnotherGroup(t *testing.T) {
	// Group is not selected and host is not assigned to any group
	layout := constant.ScreenLayoutDescription
	hostDelegate := NewHostDelegate(&layout, lo.ToPtr(""), &testutils.MockLogger{})
	require.False(t, hostDelegate.isHostMovedToAnotherGroup(""))

	// Group is not selected and host is assigned to "Group 1". Because group is not selected
	// the host is NOT in a different group
	layout = constant.ScreenLayoutDescription
	hostDelegate = NewHostDelegate(&layout, nil, &testutils.MockLogger{})
	require.False(t, hostDelegate.isHostMovedToAnotherGroup("Group 1"))

	// Group is selected and host is assigned to "Group 1"
	layout = constant.ScreenLayoutDescription
	hostDelegate = NewHostDelegate(&layout, lo.ToPtr("Group 1"), &testutils.MockLogger{})
	require.True(t, hostDelegate.isHostMovedToAnotherGroup("Group 2"))
}

// Test cases for Render function
func TestHostDelegate_Render(t *testing.T) {
	hostNoGroup := ListItemHost{Host: host.NewHost(0, "Mock Host 1", "", "localhost", "", "", "22")}
	hostWithGroup := ListItemHost{Host: host.NewHost(0, "Mock Host 2", "", "localhost", "", "", "22")}
	hostWithGroup.Group = "Group 2"

	tests := []struct {
		appStateGroup string
		listItemHost  ListItemHost
		layout        constant.ScreenLayout
		expectedDesc  string
	}{
		{
			// When group is selected, but host doesn't have group should display "[no group]" next to title
			"Group 1",
			hostNoGroup,
			constant.ScreenLayoutDescription,
			"Mock Host 1 [no group]",
		},
		{
			// When group is selected, but host doesn't have group should display "[no group]" next to title
			"Group 1",
			hostNoGroup,
			constant.ScreenLayoutCompact,
			"Mock Host 1 [no group]",
		},
		{
			// When ScreenLayoutGroup is selected should not display group as it displayed in the description
			"Group 1",
			hostNoGroup,
			constant.ScreenLayoutGroup,
			"Mock Host 1",
		},
		{
			// When group is NOT selected, but host does have group should NOT display group next to title
			"Group 1",
			hostNoGroup,
			constant.ScreenLayoutCompact,
			"Mock Host 1",
		},
		{
			// When group is selected, but host has a different group, then should display group next to title
			"Group 1",
			hostWithGroup,
			constant.ScreenLayoutCompact,
			"Mock Host 2 (Group 2)",
		},
	}

	for _, tc := range tests {
		var buf bytes.Buffer
		mockModel := NewMockListModel(false)
		mockModel.Update(tea.WindowSizeMsg{Width: 100, Height: 100}) // required, otherwise the model does not render anything
		hostDelegate := NewHostDelegate(&tc.layout, &tc.appStateGroup, &testutils.MockLogger{})
		hostDelegate.Render(&buf, mockModel.Model, 0, tc.listItemHost)
		require.Contains(t, buf.String(), tc.expectedDesc)
	}
}
