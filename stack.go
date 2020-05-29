package gqlyzer

// this file containing stack operation

import "errors"

func (l *Lexer) pop() rune {
	tail := l.parseStack[len(l.parseStack)-1]
	l.parseStack = l.parseStack[:len(l.parseStack)-1]

	return tail
}

func (l *Lexer) popFlush() (string, error) {
	result := ""
	c := l.pop()
	for c != '#' {
		if len(l.parseStack) == 0 {
			return "", errors.New("flush not found")
		}
		result = string(c) + result
		c = l.pop()
	}

	return result, nil
}

func (l *Lexer) popCond(c rune) error {
	tail := l.parseStack[len(l.parseStack)-1]
	if tail != c {
		return errors.New("invalid stack pop")
	}
	l.parseStack = l.parseStack[:len(l.parseStack)-1]

	return nil
}

func (l *Lexer) push(c rune) {
	l.parseStack = append(l.parseStack, c)
}

func (l *Lexer) pushFlush() {
	l.parseStack = append(l.parseStack, '#')
}

//func (l *Lexer) pushString(s string) {
//	for _, c := range s {
//		l.push(c)
//	}
//}
