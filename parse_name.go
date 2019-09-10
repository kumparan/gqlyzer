package gqlyzer

func (l *Lexer) parseName() string {
	var name string
	c, err := l.read()
	for isAlphabet(c) &&
		err == nil {
		name += string(c)
		l.cursor++
		c, err = l.read()
	}

	return name
}
