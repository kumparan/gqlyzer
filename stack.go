package gqlyzer

// this file containing stack operation

import "errors"

func (l *Lexer) pop() rune {
	tail := l.parseStack[len(l.parseStack)-1]
	l.parseStack = l.parseStack[:len(l.parseStack)-1]
	return tail
}

func (l *Lexer) popFlush() error {
	for l.pop() != '#' {
		if len(l.parseStack) == 0 {
			return errors.New("flush not found")
		}
	}
	return nil
}

func (l *Lexer) popCond(c rune) error {
	tail := l.parseStack[len(l.parseStack)-1]
	if rune(tail) != c {
		return errors.New("invalid stack pop")
	}
	l.parseStack = l.parseStack[:len(l.parseStack)-1]
	return nil
}

func (l *Lexer) push(c rune) {
	l.parseStack = append(l.parseStack, c)
}

func (l *Lexer) pushString(s string) {
	for _, c := range s {
		l.push(c)
	}
}

func (l *Lexer) pushFlush() {
	l.parseStack = append(l.parseStack, '#')
}
