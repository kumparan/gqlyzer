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
		// EOF here means the input was empty or all whitespace — not a real error.
		if errors.Is(err, ErrEOF) {
			err = nil
		}
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
	case '{': // anonymous operation — advance past '{' so the caller doesn't see it again
		l.cursor++
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

	// Empty input — parseOperationType returned empty op and nil err.
	// Nothing to do; return zero value cleanly.
	if len(opType) <= 0 && !isAnonymous {
		return
	}

	op.Type = opType

	if isAnonymous {
		// cursor already sits after '{'; parseSelectionSet expects to see '{',
		// so back up one position.
		l.cursor--
		op.Selections, err = l.parseSelectionSet()
		return
	}

	// Named operation: check what follows the keyword — a name, '(' (variables),
	// or '{' (no explicit name, straight to body).
	l.consumeWhitespace()
	c, err := l.read()
	if err != nil {
		// EOF immediately after keyword with no body is a real mid-parse error.
		return
	}

	if c == '{' {
		// Keyword with no name, e.g.  query { ... }
		// Do NOT advance; parseSelectionSet needs to see '{'.
		op.Selections, err = l.parseSelectionSet()
		return
	}

	if c != '(' {
		// There is a name to parse. parseName reads from the current cursor
		// position, so do NOT advance here.
		var name string
		name, err = l.parseName()
		if err != nil {
			return token.Operation{}, err
		}
		op.Name = name
		l.consumeWhitespace()

		// Re-read to find out what comes after the name.
		c, err = l.read()
		if err != nil {
			return
		}
	}

	// Skip variable definitions:  ( $var: Type, … )
	if c == '(' {
		l.cursor++
		c, err = l.read()
		for err == nil && c != ')' {
			l.cursor++
			c, err = l.read()
		}
		if err != nil {
			// EOF inside variable list is a real mid-parse error.
			return
		}
		l.cursor++ // consume ')'
		l.consumeWhitespace()
	}

	// Cursor must now be sitting on '{'.
	op.Selections, err = l.parseSelectionSet()
	return
}
