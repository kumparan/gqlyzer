package gqlyzer

import (
	"fmt"

	"github.com/kumparan/gqlyzer/token"
)

func (l *Lexer) parseSelection() (newSelection token.Selection, err error) {
	if x := l.pop(); x != ',' && x != '\\' {
		err = fmt.Errorf("expected separator, but got: %s", string(x))
		return
	}

	name, err := l.parseName()
	if err != nil {
		return
	}
	newSelection.Name = name

	arguments, argErr := l.parseArgumentSet()
	if argErr == nil && len(arguments) > 0 {
		newSelection.Arguments = arguments
		l.cursor++
		l.consumeWhitespace()
	}

	subSelection, subErr := l.parseSelectionSet()
	if subErr == nil {
		newSelection.InnerSelection = subSelection

	}

	return
}

func (l *Lexer) parseSelectionSet() (set token.SelectionSet, err error) {
	set = make(token.SelectionSet)
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
		l.cursor++
		l.consumeWhitespace()
		c, err = l.read()
		for err == nil && c != '}' {
			if c == ',' {
				l.push(c)
				l.cursor++
				l.consumeWhitespace()
				c, err = l.read()
				continue
			}

			selection, err := l.parseSelection()
			if err != nil {
				fmt.Print(set)
				return token.SelectionSet{}, err
			}
			set[selection.Name] = selection
			l.consumeWhitespace()
			c, err = l.read()
			if err != nil {
				return token.SelectionSet{}, err
			}
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

	l.cursor++
	return
}
