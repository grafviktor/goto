// Package ui - contains UI iteraction code.
package ui

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/grafviktor/goto/internal/constant"
	"github.com/grafviktor/goto/internal/model/sshconfig"
	"github.com/grafviktor/goto/internal/state"
	"github.com/grafviktor/goto/internal/storage"
	"github.com/grafviktor/goto/internal/ui/component/grouplist"
	"github.com/grafviktor/goto/internal/ui/component/hostedit"
	"github.com/grafviktor/goto/internal/ui/component/hostlist"
	"github.com/grafviktor/goto/internal/ui/message"
	"github.com/grafviktor/goto/internal/utils"
)

type iLogger interface {
	Debug(format string, args ...any)
	Info(format string, args ...any)
	Warn(format string, args ...any)
	Error(format string, args ...any)
}

// New - creates a parent module for other component and preserves state which
// can be propagated to other subcomponents.
func New(
	ctx context.Context,
	storage storage.HostStorage,
	appState *state.Application,
	log iLogger,
) mainModel {
	m := mainModel{
		modelHostList:  hostlist.New(ctx, storage, appState, log),
		modelGroupList: grouplist.New(ctx, storage, appState, log),
		appContext:     ctx,
		hostStorage:    storage,
		appState:       appState,
		logger:         log,
	}

	return m
}

type mainModel struct {
	appContext         context.Context
	hostStorage        storage.HostStorage
	modelHostList      tea.Model
	modelGroupList     tea.Model
	modelHostEdit      tea.Model
	appState           *state.Application
	viewMessageContent string
	logger             iLogger
	viewport           viewport.Model
	ready              bool
}

func (m *mainModel) Init() tea.Cmd {
	m.logger.Debug("[UI] Run init function")

	// Loads hosts from DB
	return m.modelHostList.Init()
}

func (m *mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.logger.Debug("[UI] Keyboard event: '%v'", msg)
		return m.handleKeyEvent(msg)
	case tea.WindowSizeMsg:
		m.logger.Debug("[UI] Set terminal window size: %d %d", msg.Width, msg.Height)
		m.appState.Width = msg.Width
		m.appState.Height = msg.Height
		m.updateViewPort(msg.Width, msg.Height)
	case message.OpenViewHostEdit:
		m.logger.Debug("[UI] Open host edit form")
		m.appState.CurrentView = state.ViewEditItem
		ctx := context.WithValue(m.appContext, hostedit.ItemID, msg.HostID)
		m.modelHostEdit = hostedit.New(ctx, m.hostStorage, m.appState, m.logger)
	case message.CloseViewHostEdit:
		m.logger.Debug("[UI] Close host edit form")
		m.appState.CurrentView = state.ViewHostList
	case message.OpenViewSelectGroup:
		m.logger.Debug("[UI] Open select group form")
		m.appState.CurrentView = state.ViewGroupList
	case message.CloseViewSelectGroup:
		m.logger.Debug("[UI] Close select group form")
		m.appState.CurrentView = state.ViewHostList
	case message.HostSelected:
		m.logger.Debug("[UI] Update app state. Active host id: %d", msg.HostID)
		m.appState.Selected = msg.HostID
	case message.RunProcessSSHConnect:
		m.logger.Debug("[UI] Connect to focused SSH host")
		return m, m.dispatchProcessSSHConnect(msg)
	case message.RunProcessSSHLoadConfig:
		m.logger.Debug("[UI] Load SSH config for focused host id: %d, title: %q", msg.Host.ID, msg.Host.Title)
		return m, m.dispatchProcessSSHLoadConfig(msg)
	case message.RunProcessSSHCopyID:
		m.logger.Debug("[UI] Copy SSH config to host id: %d, title: %q", msg.Host.ID, msg.Host.Title)
		return m, m.dispatchProcessSSHCopyID(msg)
	case message.RunProcessSuccess:
		m.logger.Debug("[UI] Handle process success message. Process: %v", msg.ProcessType)
		cmd = m.handleProcessSuccess(msg)
		cmds = append(cmds, cmd)
	case message.RunProcessErrorOccurred:
		m.logger.Debug("[UI] Handle process error message. Process: %v", msg.ProcessType)
		m.handleProcessError(msg)
	}

	m.modelHostList, cmd = m.modelHostList.Update(msg)
	cmds = append(cmds, cmd)
	m.modelGroupList, cmd = m.modelGroupList.Update(msg)
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
	case state.ViewHostList:
		content = m.modelHostList.View()
	case state.ViewGroupList:
		content = m.modelGroupList.View()
	case state.ViewMessage:
		content = m.viewMessageContent
	case state.ViewEditItem:
		content = m.modelHostEdit.View()
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
	case state.ViewMessage:
		// When display external process's output and receive any keyboard event, we:
		// 1. Reset the error message
		// 2. Switch to HostList view
		m.viewMessageContent = ""
		m.appState.CurrentView = state.ViewHostList
	case state.ViewHostList:
		m.modelHostList, cmd = m.modelHostList.Update(msg)
	case state.ViewGroupList:
		m.modelGroupList, cmd = m.modelGroupList.Update(msg)
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

func (m *mainModel) dispatchProcess(
	processType constant.ProcessType,
	process *exec.Cmd,
	inBackground,
	ignoreError bool,
) tea.Cmd {
	onProcessExitCallback := func(err error) tea.Msg {
		// We can only read StdOut or StdErr of a process which was built using `BuildProcessInterceptStdAll()`
		// function because it preserves process output in a temporary buffer.
		var processOutput string
		if readableStdOut, ok := process.Stdout.(*utils.ProcessBufferWriter); ok {
			processOutput = strings.TrimSpace(string(readableStdOut.Output))
		}

		var readableStdErr string
		if readableErrOutput, ok := process.Stderr.(*utils.ProcessBufferWriter); ok {
			readableStdErr = strings.TrimSpace(string(readableErrOutput.Output))
		}

		// This callback triggers when external process exits
		if err != nil {
			if utils.StringEmpty(&readableStdErr) {
				readableStdErr = err.Error()
			}

			// Sometimes we don't care when external process ends with an error.
			if ignoreError {
				m.logger.Info("[EXEC] Terminate process with reason %v. Error ignored.", readableStdErr)
				return nil
			}

			m.logger.Error("[EXEC] Terminate process with reason %v", readableStdErr)
			commandWhichFailed := strings.Join(process.Args, " ")
			// errorDetails contains command which was executed and the error text.
			errorDetails := fmt.Sprintf("Command: %s\nError:   %s", commandWhichFailed, readableStdErr)
			return message.RunProcessErrorOccurred{
				ProcessType: processType,
				StdOut:      processOutput,
				StdErr:      errorDetails,
			}
		}

		m.logger.Info("[EXEC] Terminate process gracefully: %s", process.String())

		return message.RunProcessSuccess{
			ProcessType: processType,
			StdOut:      processOutput,
			StdErr:      readableStdErr,
		}
	}

	if inBackground {
		// If process runs in background we have to read its output and store in msg.
		return func() tea.Msg {
			err := process.Run()

			return onProcessExitCallback(err)
		}
	}

	// tea.ExecProcess always runs in a foreground.
	// Return value is 'tea.Cmd' struct
	return tea.ExecProcess(process, onProcessExitCallback)
}

func (m *mainModel) dispatchProcessSSHConnect(msg message.RunProcessSSHConnect) tea.Cmd {
	m.logger.Debug("[EXEC] Build ssh connect command for hostname: %v, title: %v", msg.Host.Address, msg.Host.Title)
	process := utils.BuildProcessInterceptStdErr(msg.Host.CmdSSHConnect())
	m.logger.Info("[EXEC] Run process: '%s'", process.String())

	return m.dispatchProcess(constant.ProcessTypeSSHConnect, process, false, false)
}

func (m *mainModel) dispatchProcessSSHLoadConfig(msg message.RunProcessSSHLoadConfig) tea.Cmd {
	m.logger.Debug("[EXEC] Read ssh configuration for host: '%+v'", msg.Host)
	process := utils.BuildProcessInterceptStdAll(msg.Host.CmdSSHConfig())
	m.logger.Info("[EXEC] Run process: '%s'", process.String())

	// Should run in non-blocking fashion for ssh load config
	return m.dispatchProcess(constant.ProcessTypeSSHLoadConfig, process, true, true)
}

func (m *mainModel) dispatchProcessSSHCopyID(msg message.RunProcessSSHCopyID) tea.Cmd {
	identityFile, hostname := msg.Host.SSHHostConfig.IdentityFile, msg.Host.SSHHostConfig.Hostname
	m.logger.Debug("[EXEC] Copy ssh-key '%s.pub' to host '%s'", identityFile, hostname)
	if sshconfig.IsAlternativeFilePathDefined() {
		m.logger.Warn("[EXEC] copy ssh key when alternative ssh config file is used: %q. ssh config file is ignored.",
			m.appState.ApplicationConfig.SSHConfigFilePath)
	}
	process := utils.BuildProcessInterceptStdAll(msg.Host.CmdSSHCopyID())
	m.logger.Info("[EXEC] Run process: '%s'", process.String())

	// Should run in non-blocking fashion for ssh copy id
	return m.dispatchProcess(constant.ProcessTypeSSHCopyID, process, false, false)
}

func (m *mainModel) handleProcessSuccess(msg message.RunProcessSuccess) tea.Cmd {
	if msg.ProcessType == constant.ProcessTypeSSHLoadConfig {
		parsedSSHConfig := sshconfig.Parse(msg.StdOut)
		m.logger.Debug("[EXEC] Host SSH config loaded: %+v", *parsedSSHConfig)
		return message.TeaCmd(message.HostSSHConfigLoaded{
			HostID: m.appState.Selected,
			Config: *parsedSSHConfig,
		})
	}

	if msg.ProcessType == constant.ProcessTypeSSHCopyID {
		m.logger.Debug("[EXEC] Host SSH key copied. Details:\n%s\n%s", msg.StdOut, msg.StdErr)

		if strings.Contains(msg.StdErr, "ERROR") || strings.Contains(msg.StdErr, "WARNING") {
			m.viewMessageContent = msg.StdErr
		} else {
			m.viewMessageContent = msg.StdOut
		}

		m.appState.CurrentView = state.ViewMessage
	}

	return nil
}

func (m *mainModel) handleProcessError(msg message.RunProcessErrorOccurred) {
	var errMsg string
	if !utils.StringEmpty(&msg.StdOut) {
		errMsg = fmt.Sprintf("%s\nDetails: %s", msg.StdErr, msg.StdOut)
	} else {
		errMsg = msg.StdErr
	}

	// We use m.logger.Debug method to report about the error,
	// because the error was already reported by run process module.
	m.logger.Debug("[EXEC] External process error. %v", errMsg)
	m.viewMessageContent = errMsg
	m.appState.CurrentView = state.ViewMessage
}
