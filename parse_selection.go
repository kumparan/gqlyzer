package gqlyzer

import (
	"github.com/kumparan/gqlyzer/token"
)

func (l *Lexer) parseSelection() (newSelection token.Selection, err error) {
	name, err := l.parseName()
	if err != nil {
		return
	}
	newSelection.Name = name

	subSelection, subErr := l.parseSelectionSet()
	if subErr == nil {
		newSelection.InnerSelection = subSelection
	}

	return
}

func (l *Lexer) parseSelectionSet() (set token.SelectionSet, err error) {
	l.consumeWhitespace()
	c, err := l.read()
	if err != nil {
		return
	}

	if c == '{' {
		l.push('{')
		l.pushFlush()
		if err != nil {
			return
		}

		l.consumeWhitespace()
		c, err = l.read()
		for err == nil && c != '}' {
			l.parseSelection()
			l.consumeWhitespace()
			l.cursor++
			c, err = l.read()
		}

		_, err = l.popFlush()
		if err != nil {
			return
		}

		err = l.popCond('{')
		if err != nil {
			return
		}
	} else {
		return token.SelectionSet{}, nil
	}

	return
}
