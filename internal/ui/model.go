// Package ui - contains UI iteraction code.
package ui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/grafviktor/goto/internal/state"
	"github.com/grafviktor/goto/internal/storage"
	"github.com/grafviktor/goto/internal/ui/component/hostedit"
	"github.com/grafviktor/goto/internal/ui/component/hostlist"
	"github.com/grafviktor/goto/internal/ui/message"
	"github.com/grafviktor/goto/internal/utils"
)

type logger interface {
	Debug(format string, args ...any)
	Info(format string, args ...any)
	Error(format string, args ...any)
}

// New - creates a parent module for other component and preserves state which
// can be propagated to other sub-components.
func New(
	ctx context.Context,
	storage storage.HostStorage,
	appState *state.ApplicationState,
	log logger,
) mainModel {
	m := mainModel{
		modelHostList: hostlist.New(ctx, storage, appState, log),
		appContext:    ctx,
		hostStorage:   storage,
		appState:      appState,
		logger:        log,
	}

	return m
}

type mainModel struct {
	appContext    context.Context
	hostStorage   storage.HostStorage
	modelHostList tea.Model
	modelHostEdit tea.Model
	appState      *state.ApplicationState
	logger        logger
	viewport      viewport.Model
	ready         bool
}

func (m *mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.logger.Debug("[UI] Keyboard event: %v", msg)
		return m.handleKeyEvent(msg)
	case message.TerminalSizePolling:
		// That is Windows OS specific. Windows cmd.exe does not trigger terminal
		// resize events, that is why we poll terminal size with intervals
		// First message is being triggered by Windows version of the model.
		if msg.Width != m.appState.Width || msg.Height != m.appState.Height {
			m.logger.Debug("[UI] Terminal size polling message received: %d %d", msg.Width, msg.Height)
			cmds = append(cmds, message.TeaCmd(tea.WindowSizeMsg{Width: msg.Width, Height: msg.Height}))
		}

		// We dispatch the same message from this function and therefore cycle TerminalSizePollingMsg.
		// That's done on purpose to keep this process running. Message.TerminalSizePollingMsg will trigger
		// automatically after an artificial delay which is set by Time.Sleep inside message.
		cmds = append(cmds, message.TerminalSizePollingMsg)
	case tea.WindowSizeMsg:
		m.logger.Debug("[UI] Set terminal window size: %d %d", msg.Width, msg.Height)
		m.appState.Width = msg.Width
		m.appState.Height = msg.Height
		m.updateViewPort(msg.Width, msg.Height)
	case hostlist.OpenEditForm:
		m.logger.Debug("[UI] Open host edit form")
		m.appState.CurrentView = state.ViewEditItem
		ctx := context.WithValue(m.appContext, hostedit.ItemID, msg.HostID)
		m.modelHostEdit = hostedit.New(ctx, m.hostStorage, m.appState, m.logger)
	case message.HostListSelectItem:
		m.logger.Debug("[UI] Update app state. Active host id: %d", msg.HostID)
		m.appState.Selected = msg.HostID
	case hostedit.MsgClose:
		m.logger.Debug("[UI] Close host edit form")
		m.appState.CurrentView = state.ViewHostList
	case message.RunProcessConnectSSH:
		return m, m.dispatchProcessConnect(msg)
	case message.RunProcessLoadHostConfig:
		return m, m.dispatchProcessLoadConfig(msg)
	case message.RunProcessErrorOccurred:
		// We use m.logger.Debug method to report about the error,
		// because the error was already reported by run process module.
		m.logger.Debug("[UI] External process error. %v", msg.Err)
		m.appState.Err = msg.Err
		m.appState.CurrentView = state.ViewErrorMessage
	}

	m.modelHostList, cmd = m.modelHostList.Update(msg)
	cmds = append(cmds, cmd)

	if m.appState.CurrentView == state.ViewEditItem {
		// Edit host receives messages only if it's active. We re-create this component every time we go to edit mode
		m.modelHostEdit, cmd = m.modelHostEdit.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *mainModel) View() string {
	// Build UI
	var content string
	switch m.appState.CurrentView {
	case state.ViewErrorMessage:
		content = m.appState.Err.Error()
	case state.ViewEditItem:
		content = m.modelHostEdit.View()
	case state.ViewHostList:
		content = m.modelHostList.View()
	}

	// Wrap UI into the ViewPort
	m.viewport.SetContent(content)
	viewPortContent := m.viewport.View()

	return viewPortContent
}

func (m *mainModel) handleKeyEvent(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyCtrlC {
		m.logger.Debug("[UI] Receive Ctrl+C. Quit the application")
		return m, tea.Quit
	}

	var cmd tea.Cmd

	// Only current view receives key messages
	switch m.appState.CurrentView {
	case state.ViewErrorMessage:
		// When display external process's output and receieve any keyboard event, we:
		// 1. Reset the error message
		// 2. Switch to HostList view
		m.appState.Err = nil
		m.appState.CurrentView = state.ViewHostList
	case state.ViewHostList:
		m.modelHostList, cmd = m.modelHostList.Update(msg)
	case state.ViewEditItem:
		m.modelHostEdit, cmd = m.modelHostEdit.Update(msg)
	}

	return m, cmd
}

func (m *mainModel) updateViewPort(w, h int) tea.Model {
	if !m.ready {
		m.ready = true
		m.viewport = viewport.New(m.appState.Width, m.appState.Height)
	} else {
		m.viewport.Width = w
		m.viewport.Height = h
	}

	return m
}

func (m *mainModel) dispatchProcess(process *exec.Cmd, errorWriter *stdErrorWriter) tea.Cmd {
	onProcessExitCallback := func(err error) tea.Msg {
		// This callback triggers when external process exits
		if err != nil {
			errorMessage := strings.TrimSpace(string(errorWriter.err))
			if utils.StringEmpty(errorMessage) {
				errorMessage = err.Error()
			}

			m.logger.Error("[EXEC] Terminate process with reason %v", errorMessage)
			commandWhichFailed := strings.Join(process.Args, " ")
			// errorDetails contains command which was executed and the error text.
			errorDetails := fmt.Sprintf("Command: %s\nError:   %s", commandWhichFailed, errorMessage)
			return message.RunProcessErrorOccurred{Err: errors.New(errorDetails)}
		}

		m.logger.Info("[EXEC] Terminate process gracefully: %s", process.String())
		return nil
	}

	// Return value is 'tea.Cmd' struct
	return tea.ExecProcess(process, onProcessExitCallback)
}

func (m *mainModel) dispatchProcessConnect(msg message.RunProcessConnectSSH) tea.Cmd {
	var process *exec.Cmd
	errorWriter := stdErrorWriter{}
	m.logger.Debug("[EXEC] Build ssh connect command for hostname: %v, title: ", msg.Host.Address, msg.Host.Title)
	process = utils.BuildConnectSSH(msg.Host, &errorWriter)
	m.logger.Info("[EXEC] Run process: %s", process.String())

	return m.dispatchProcess(process, &errorWriter)
}

func (m *mainModel) dispatchProcessLoadConfig(msg message.RunProcessLoadHostConfig) tea.Cmd {
	var process *exec.Cmd
	errorWriter := stdErrorWriter{}
	m.logger.Debug("[EXEC] Read ssh configuration for hostname: %v, title: ", msg.Hostname)
	process = utils.BuildLoadSSHConfig(msg.Hostname, os.Stdout, &errorWriter)

	return m.dispatchProcess(process, &errorWriter)
}

// stdErrorWriter - is an object which pretends to be a writer, however it saves all data into 'err' variable
// for future reading and do not write anything in terminal. We need it to display a formatted error in the console
// when it's required, but not when it's done by default.
type stdErrorWriter struct {
	err []byte
}

// Write - doesn't write anything, it saves all data in err variable, which can ve read later.
func (writer *stdErrorWriter) Write(p []byte) (n int, err error) {
	writer.err = append(writer.err, p...)

	// Hide error from the console, otherwise it will be seen in a subsequent ssh calls
	// To return to default behavior use: return os.Stderr.Write(p)
	// We must return the number of bytes which were written using `len(p)`,
	// otherwise exec.go will throw 'short write' error.
	return len(p), nil
}
