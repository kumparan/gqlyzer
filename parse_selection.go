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

	if l.isEOF() {
		newSelection.Name = name
		return
	}

	c, err := l.read()
	if err != nil {
		return
	}

	if c == ':' {
		newSelection.Alias = name
		l.cursor++
		l.consumeWhitespace()
		name, err = l.parseName()
		if err != nil {
			return
		}
		newSelection.Name = name
	} else {
		newSelection.Name = name
	}

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

			switch {
			case l.parseStack[len(l.parseStack)-1] == ',':
				goto parseSelection
			case l.parseStack[len(l.parseStack)-1] != '\\': // TODO: cases with space separator i.e. directives (currently ignored)
				for err == nil && c != '\n' {
					l.cursor++
					c, err = l.read()
					continue
				}
				l.consumeWhitespace()
			}

		parseSelection:
			selection, err := l.parseSelection()
			if err != nil {
				return set, err // TODO: currently only need selection name so errors bypassed, handle more complex cases such as directives
			}
			set[selection.Name] = selection
			l.consumeWhitespace()
			c, err = l.read()
			if err != nil {
				return set, err // TODO: currently only need selection name so errors bypassed, handle more complex cases such as directives
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
		return set, nil
	}

	l.cursor++
	return
}
