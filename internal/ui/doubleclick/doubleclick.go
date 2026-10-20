package doubleclick

import (
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
)

type Callback func() tea.Cmd

const timeToWaitForSecondClick = 300 * time.Millisecond

type secondClickTimeOut struct{ componentName string }

type DoubleClick struct {
	componentName      string
	waitForSecondClick bool
	firstClick         Callback
	secondClick        Callback
	listModel          *list.Model
}

func New(
	owningComponentName string,
	listModel *list.Model,
	firstClick Callback,
	secondClick Callback,
) *DoubleClick {
	return &DoubleClick{
		componentName: owningComponentName,
		listModel:     listModel,
		firstClick:    firstClick,
		secondClick:   secondClick,
	}
}

func (d *DoubleClick) HandleClick(index int) tea.Cmd {
	if index < 0 {
		return nil
	}
	if d.waitForSecondClick && index == d.listModel.Index() {
		d.waitForSecondClick = false
		return d.secondClick()
	}
	d.listModel.Select(index)
	d.waitForSecondClick = true
	return tea.Batch(
		d.firstClick(),
		tea.Tick(timeToWaitForSecondClick, func(time.Time) tea.Msg {
			return secondClickTimeOut{d.componentName}
		}),
	)
}
