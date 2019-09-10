package gqlyzer

import "errors"

func (l *Lexer) parseKeyword(keyword string) error {
	l.pushFlush()
	c, err := l.read()
	for isAlphabet(c) && err == nil {
		l.push(c)
		l.cursor++
		c, err = l.read()
	}

	k, err := l.popFlush()

	if k != keyword {
		return errors.New("unknown keyword: " + k)
	}

	return err
}
