package gqlyzer

import (
	"errors"
)

func (l *Lexer) parseName() (string, error) {
	var name string
	c, err := l.read()
	if !isAlphabet(c) && c != '_' {
		return "", errors.New("first character of an identifier Name have to be an alphabet or underscore: " + string(c))
	}

	for (isAlphabet(c) || c == '_' || isNumber(c)) &&
		err == nil {
		name += string(c)
		l.cursor++
		c, err = l.read()
	}

	return name, nil
}
