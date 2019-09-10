package gqlyzer

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
