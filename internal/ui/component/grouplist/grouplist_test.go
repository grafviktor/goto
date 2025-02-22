package grouplist

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/grafviktor/goto/internal/state"
	"github.com/grafviktor/goto/internal/test"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	gm := NewMockGroupModel(false)
	assert.False(t, gm.FilteringEnabled())
	assert.False(t, gm.ShowStatusBar())
	assert.False(t, gm.FilteringEnabled())
	assert.Equal(t, gm.Title, "select group")
}

func TestInit(t *testing.T) {
	gm := NewMockGroupModel(false)
	cmd := gm.Init()
	assert.IsType(t, tea.Cmd(nil), cmd)
}

// func TestUpdate(t *testing.T) {
// 	gm := NewMockGroupModel(false)
// 	cmd := gm.Update()
// 	assert.IsType(t, tea.Cmd(nil), cmd)
// }

// ==============================================
// ============== utility methods ===============
// ==============================================

func NewMockGroupModel(storageShouldFail bool) *model {
	mockState := state.ApplicationState{Selected: 1}
	storage := test.NewMockStorage(storageShouldFail)
	return New(context.TODO(), storage, &mockState, &test.MockLogger{})
}
