package sshconfig

import (
	"errors"
	"strings"

	model "github.com/grafviktor/goto/internal/model/host"
	"github.com/grafviktor/goto/internal/utils"
)

type lexer interface {
	Tokenize() ([]SSHToken, error)
}

// Parser is responsible for parsing SSH configuration tokens into Host models.
type Parser struct {
	lexer       lexer
	currentHost *model.Host
	foundHosts  []model.Host
	logger      iLogger
}

// NewParser constructs a new Parser instance with the provided lexer and logger.
func NewParser(lexer lexer, log iLogger) *Parser {
	return &Parser{
		lexer:  lexer,
		logger: log,
	}
}

// Parse processes the tokens from the lexer and constructs a slice of Host models.
func (p *Parser) Parse() ([]model.Host, error) {
	if p.lexer == nil {
		return nil, errors.New("lexer is not set")
	}

	hostTokens, err := p.lexer.Tokenize()
	if err != nil {
		return nil, err
	}
	p.currentHost = nil
	p.foundHosts = nil

	for _, token := range hostTokens {
		switch token.kind {
		case tokenKind.Host:
			// New host found, append current host if it is valid.
			p.appendLastHostIfValid()
			p.currentHost = &model.Host{
				Title: token.value,
			}
		case tokenKind.Hostname:
			p.currentHost.Address = token.value
		case tokenKind.NetworkPort:
			p.currentHost.RemotePort = token.value
		case tokenKind.IdentityFile:
			p.currentHost.IdentityFilePath = token.value
		case tokenKind.User:
			p.currentHost.LoginName = token.value
		case tokenKind.Group:
			p.currentHost.Group = token.value
		case tokenKind.Description:
			p.currentHost.Description = token.value
		}
	}

	p.appendLastHostIfValid()
	p.setDefaults()

	return p.foundHosts, nil
}

func (p *Parser) appendLastHostIfValid() {
	if p.hostValid() {
		p.foundHosts = append(p.foundHosts, *p.currentHost)
	}
}

func (p *Parser) hostValid() bool {
	if p.currentHost == nil {
		return false
	}

	if strings.Contains(p.currentHost.Title, "*") {
		return false
	}

	if utils.StringEmpty(&p.currentHost.Title) && utils.StringEmpty(&p.currentHost.Address) {
		return false
	}

	return true
}

const putSSHConfigHostsIntoGroupName = "ssh_config"

func (p *Parser) setDefaults() {
	for i, host := range p.foundHosts {
		if utils.StringEmpty(&host.Group) {
			p.foundHosts[i].Group = putSSHConfigHostsIntoGroupName
		}

		// In ssh_config, it is valid to define a host without an address.
		// If the address is empty, then title must be used.
		if utils.StringEmpty(&host.Address) {
			p.foundHosts[i].Address = host.Title
		}
	}
}
