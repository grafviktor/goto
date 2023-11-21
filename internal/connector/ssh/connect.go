package ssh

import (
	"os/exec"
	"strings"

	"github.com/grafviktor/goto/internal/model"
)

const (
	optionPrivateKey = "-i"
	optionRemotePort = "-p"
	optionLoginName  = "-l"
)

func ConnectCmd(h model.Host) *exec.Cmd {
	sshCmd := SSHCmd()
	args := []string{}

	privateKeyPath := strings.Trim(h.PrivateKeyPath, " ")
	if privateKeyPath != "" {
		args = append(args, optionPrivateKey)
		args = append(args, privateKeyPath)
	}

	remotePort := strings.Trim(h.RemotePort, " ")
	if remotePort != "" {
		args = append(args, optionRemotePort)
		args = append(args, remotePort)
	}

	loginName := strings.Trim(h.LoginName, " ")
	if loginName != "" {
		args = append(args, optionLoginName)
		args = append(args, loginName)
	}

	args = append(sshCmd[1:], args...)

	return exec.Command(sshCmd[0], append(args, h.Address)...)
}
