package gqlyzer

import (
	"errors"

	"github.com/kumparan/gqlyzer/token"
	"github.com/kumparan/gqlyzer/token/operation"
)

func (l *Lexer) parseOperationType() (op operation.Type, isAnonymous bool, err error) {
	l.consumeWhitespace()
	l.pushFlush()

	c, err := l.read()
	if err != nil {
		return
	}
	switch c {
	case 'q':
		if err = l.parseKeyword("query"); err != nil {
			return
		}
		return operation.Query, false, nil
	case 'm':
		if err = l.parseKeyword("mutation"); err != nil {
			return
		}
		return operation.Mutation, false, nil
	case 's':
		if err = l.parseKeyword("subscription"); err != nil {
			return
		}
		return operation.Subscription, false, nil
	case '{': // anonymous operation returns query type
		return operation.Query, true, nil
	default: // TODO: parse fragments
		err = errors.New("unknown definition")
		return
	}
}

func (l *Lexer) parseOperation() (op token.Operation, err error) {
	opType, isAnonymous, err := l.parseOperationType()
	if err != nil {
		return
	}

	op.Type = opType

	// check if query is json anonymous but uses variables
	l.consumeWhitespace()
	c, err := l.read()
	if err != nil {
		return
	}

	if c == '(' || c == '{' && !isAnonymous { // check anonymous json query doesnt use variable
		isAnonymous = true
	}

	if !isAnonymous {
		name, err := l.parseName()
		if err != nil {
			return token.Operation{}, err
		}
		op.Name = name
	}

	c, err = l.read()
	if err != nil {
		return
	}

	// ignore variable of operation
	if c == '(' {
		l.cursor++
		c, err = l.read()
		for err == nil && c != ')' {
			l.cursor++
			c, err = l.read()
		}

		if err != nil {
			return
		}

		l.cursor++
		l.consumeWhitespace()
	}

	op.Selections, err = l.parseSelectionSet()
	if err != nil {
		return
	}

	return
}
