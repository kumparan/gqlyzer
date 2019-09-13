package gqlyzer

import (
	"errors"

	"github.com/kumparan/gqlyzer/token"
	"github.com/kumparan/gqlyzer/token/operation"
)

func (l *Lexer) parseOperation() (op token.Operation, err error) {
	l.consumeWhitespace()
	l.pushFlush()
	// parse "query" keyword
	var (
		isQuery    bool
		isMutation bool
	)

	c, err := l.read()
	if err != nil {
		return
	}
	switch c {
	case 'q':
		if err = l.parseKeyword("query"); err != nil {
			return
		}
		isQuery = true
		l.cursor++
	case 'm':
		if err = l.parseKeyword("mutation"); err != nil {
			return
		}
		isMutation = true
		l.cursor++
	case '{':
		break
	default:
		err = errors.New("unknown definition")
		return
	}

	if isQuery {
		op.Type = operation.Query
	} else if isMutation {
		op.Type = operation.Mutation
	}

	// get name of named operation
	if isQuery || isMutation {
		op.Name, err = l.parseName()
		if err != nil {
			return
		}
		l.pushString(op.Name)
	} else {
		op.Type = operation.Query
	}

	op.Selections, err = l.parseSelectionSet()
	if err != nil {
		return
	}

	return
}
