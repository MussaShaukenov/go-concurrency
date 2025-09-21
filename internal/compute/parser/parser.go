package parser

import (
	"errors"
	"regexp"
	"strings"

	"go.uber.org/zap"
)

// constants for commands
const (
	SetCommand = "SET"
	GetCommand = "GET"
	DelCommand = "DEL"
)

const (
	SetCommandArgCount = 3
	GetCommandArgCount = 2
	DelCommandArgCount = 2
)

var (
	ErrWrongLength     = errors.New("wrong length of command")
	ErrInvalidCommand  = errors.New("invalid command")
	ErrInvalidArgument = errors.New("invalid argument format")
)

// Regex for /(\w+)/g pattern
// Allows letters, digits, underscores, asterisks, and forward slashes
var argumentRegex = regexp.MustCompile(`^[\w*/_]+$`)

type Engine interface {
	Set(key, value any)
	Get(key any) (any, error)
	Delete(key any) error
}

type Parser struct {
	log    *zap.SugaredLogger
	engine Engine
}

func NewParser(log *zap.SugaredLogger, engine Engine) *Parser {
	return &Parser{
		log:    log,
		engine: engine,
	}
}

func (p *Parser) ParseQuery(in string) (any, error) {
	p.log.Infof("processing query: %s", strings.TrimSpace(in))

	query := strings.Fields(in)
	if len(query) == 0 {
		return nil, ErrInvalidCommand
	}

	// Validate command structure
	err := p.validateCommand(query)
	if err != nil {
		p.log.Errorf("command validation failed: %v", err)
		return nil, err
	}

	// Validate argument format
	if err := p.validateArguments(query[1:]); err != nil {
		p.log.Errorf("argument validation failed: %v", err)
		return nil, err
	}

	// Execute command
	result := p.ExecuteQuery(query)
	p.log.Info("query processed successfully")

	return result, nil
}

func (p *Parser) validateCommand(query []string) error {
	command := query[0]

	switch command {
	case SetCommand:
		if len(query) != SetCommandArgCount {
			return ErrWrongLength
		}
	case GetCommand:
		if len(query) != GetCommandArgCount {
			return ErrWrongLength
		}
	case DelCommand:
		if len(query) != DelCommandArgCount {
			return ErrWrongLength
		}
	default:
		return ErrInvalidCommand
	}

	return nil
}

func (p *Parser) validateArguments(args []string) error {
	for _, arg := range args {
		if !argumentRegex.MatchString(arg) {
			return ErrInvalidArgument
		}
	}
	return nil
}

func (p *Parser) ExecuteQuery(query []string) any {
	command := query[0]

	switch command {
	case SetCommand:
		p.Set(query[1], query[2])
		return nil
	case GetCommand:
		result, err := p.Get(query[1])
		if err != nil {
			return err.Error()
		}
		return result
	case DelCommand:
		err := p.Del(query[1])
		if err != nil {
			return err.Error()
		}
		return nil
	}

	return nil
}

func (p *Parser) Set(key, value any) {
	p.engine.Set(key, value)
}

func (p *Parser) Get(key any) (any, error) {
	value, err := p.engine.Get(key)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (p *Parser) Del(key any) error {
	return p.engine.Delete(key)
}
