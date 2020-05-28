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
		isQuery, isMutation, isSubscription bool
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
	case 's':
		if err = l.parseKeyword("subscription"); err != nil {
			return
		}
		isSubscription = true
		l.cursor++
	case '{':
		break
	default:
		err = errors.New("unknown definition")
		return
	}

	switch true {
	case isQuery:
		op.Type = operation.Query
	case isSubscription:
		op.Type = operation.Subscription
	case isMutation:
		op.Type = operation.Mutation
	}

	// get name of named operation
	if isQuery || isMutation || isSubscription {
		name, err := l.parseName()
		if err != nil {
			return token.Operation{}, err
		}
		op.Name = name
	} else {
		op.Type = operation.Query
	}

	l.consumeWhitespace()
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
