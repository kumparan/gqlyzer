package gqlyzer

import (
	"errors"
)

func isNumber(c rune) bool {
	if c >= '0' && c <= '9' {
		return true
	}

	return false
}

func isAlphabet(c rune) bool {
	if (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c < 'Z') {
		return true
	}

	return false
}

func isWhitespace(c rune) bool {
	switch c {
	case '\n', '\t', ' ':
		return true
	default:
		return false
	}
}

func (l *Lexer) isEOF() bool {
	return l.cursor >= len(l.input)
}

func (l *Lexer) read() (c rune, err error) {
	if l.isEOF() {
		err = errors.New("end of file")
	} else {
		c = rune(l.input[l.cursor])
	}
	return
}

func (l *Lexer) consumeWhitespace() {
	c, err := l.read()
	for err == nil && isWhitespace(c) {
		if c == '\n' {
			l.push('\\')
		}
		l.cursor++
		c, err = l.read()
	}
}

// commented out since no usage, needed in development
//func (l *Lexer) printParseStack() {
//	for _, c := range l.parseStack {
//		fmt.Print(string(c))
//	}
//	fmt.Println()
//}
