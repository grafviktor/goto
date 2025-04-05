package sshconfig

import (
	"errors"
	"strings"

	model "github.com/grafviktor/goto/internal/model/host"
	"github.com/grafviktor/goto/internal/utils"
)

type Lexer interface {
	Tokenize() []Token
}

type Parser struct {
	lexer       Lexer
	currentHost *model.Host
	foundHosts  []model.Host
	logger      iLogger
}

func NewParser(lexer Lexer, log iLogger) *Parser {
	return &Parser{
		lexer:  lexer,
		logger: log,
	}
}

func (p *Parser) Parse() ([]model.Host, error) {
	if p.lexer == nil {
		return nil, errors.New("Lexer is not set")
	}

	hostTokens := p.lexer.Tokenize()
	p.currentHost = nil
	p.foundHosts = nil

	for _, token := range hostTokens {
		switch token.Type {
		case TokenType.HOST:
			// New host found, append current host if it is valid.
			p.appendLastHostIfValid()
			p.currentHost = &model.Host{
				Title: token.Value(),
			}
		case TokenType.HOSTNAME:
			p.currentHost.Address = token.Value()
		case TokenType.NETWORK_PORT:
			p.currentHost.RemotePort = token.Value()
		case TokenType.IDENTITY_FILE:
			p.currentHost.IdentityFilePath = token.Value()
		case TokenType.USER:
			p.currentHost.LoginName = token.Value()
		case TokenType.GROUP:
			p.currentHost.Group = token.Value()
		case TokenType.DESCRIPTION:
			p.currentHost.Description = token.Value()
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

	if strings.TrimSpace(p.currentHost.Title) == "" || strings.Contains(p.currentHost.Title, "*") {
		return false
	}

	if strings.TrimSpace(p.currentHost.Address) == "" {
		return false
	}

	return true
}

const PUT_SSH_CONFIG_HOSTS_INTO_GROUP_NAME = "ssh_config"

func (p *Parser) setDefaults() {
	for i, host := range p.foundHosts {
		if utils.StringEmpty(&host.Group) {
			p.foundHosts[i].Group = PUT_SSH_CONFIG_HOSTS_INTO_GROUP_NAME
		}
	}
}
