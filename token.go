package filterql

import (
	"fmt"
	"strings"
	"unicode"
)

const (
	TOKEN_NONE = iota
	TOKEN_INT
	TOKEN_STR
	TOKEN_AND
	TOKEN_OR
	TOKEN_NOT
	TOKEN_ID
	TOKEN_LEFT_BRACKET
	TOKEN_RIGHT_BRACKET
	TOKEN_COMMA
	TOKEN_OP_EQ
	TOKEN_OP_NE
	TOKEN_OP_GT
	TOKEN_OP_GE
	TOKEN_OP_LT
	TOKEN_OP_LE
	TOKEN_OP_IN
	TOKEN_EOF
)

func tokenName(t int) string {
	switch t {
	case TOKEN_NONE:
		return "TOKEN_NONE"
	case TOKEN_INT:
		return "TOKEN_INT"
	case TOKEN_STR:
		return "TOKEN_STR"
	case TOKEN_AND:
		return "TOKEN_AND"
	case TOKEN_OR:
		return "TOKEN_OR"
	case TOKEN_NOT:
		return "TOKEN_NOT"
	case TOKEN_ID:
		return "TOKEN_ID"
	case TOKEN_LEFT_BRACKET:
		return "TOKEN_LEFT_BRACKET"
	case TOKEN_RIGHT_BRACKET:
		return "TOKEN_RIGHT_BRACKET"
	case TOKEN_COMMA:
		return "TOKEN_COMMA"
	case TOKEN_OP_EQ:
		return "TOKEN_OP_EQ"
	case TOKEN_OP_NE:
		return "TOKEN_OP_NE"
	case TOKEN_OP_GT:
		return "TOKEN_OP_GT"
	case TOKEN_OP_GE:
		return "TOKEN_OP_GE"
	case TOKEN_OP_LT:
		return "TOKEN_OP_LT"
	case TOKEN_OP_LE:
		return "TOKEN_OP_LE"
	case TOKEN_OP_IN:
		return "TOKEN_OP_IN"
	case TOKEN_EOF:
		return "TOKEN_EOF"
	default:
		return fmt.Sprintf("Unknown token %d", t)
	}
}

type TokenInfo struct {
	Type   int
	Text   []rune
	Offset int
}

func (ti TokenInfo) String() string {
	return fmt.Sprintf("[%s:%s@%d]", tokenName(ti.Type), string(ti.Text), ti.Offset)
}

type TokenStream struct {
	Current TokenInfo
	chars   []rune
	index   int
}

func NewTokenStream(code string) *TokenStream {
	return &TokenStream{
		chars: []rune(code),
		index: 0,
	}
}

func (ts *TokenStream) Next() bool {
	ts.skipSpaces()
	if ts.reachEnd() {
		ts.setCurrent(TOKEN_EOF, ts.index)
		return false
	}
	begin := ts.index
	ch := ts.ch()
	if ts.nextSimple(begin, ch) {
		return true
	}
	if ts.isIdChars(ch, true) {
		return ts.nextID(begin)
	}
	if ch == '\'' {
		return ts.nextStr(begin)
	}
	if ch >= '0' && ch <= '9' {
		return ts.nextInt(begin)
	}
	return true
}

func (ts *TokenStream) ch() rune {
	if ts.reachEnd() {
		return rune(0)
	}
	return ts.chars[ts.index]
}

func (ts *TokenStream) getText(begin int) []rune {
	return ts.chars[begin:ts.index]
}

func (ts *TokenStream) nextSimple(begin int, ch rune) bool {
	switch ch {
	case '=':
		ts.index++
		ts.setCurrent(TOKEN_OP_EQ, begin)
		return true
	case '(':
		ts.index++
		ts.setCurrent(TOKEN_LEFT_BRACKET, begin)
		return true
	case ')':
		ts.index++
		ts.setCurrent(TOKEN_RIGHT_BRACKET, begin)
		return true
	case ',':
		ts.index++
		ts.setCurrent(TOKEN_COMMA, begin)
		return true
	case '<':
		ts.index++
		if ts.reachEnd() {
			ts.setCurrent(TOKEN_OP_LT, begin)
		} else if ts.chars[ts.index] == '=' {
			ts.index++
			ts.setCurrent(TOKEN_OP_LE, begin)
		} else if ts.chars[ts.index] == '>' {
			ts.index++
			ts.setCurrent(TOKEN_OP_NE, begin)
		} else {
			ts.setCurrent(TOKEN_OP_LT, begin)
		}
		return true
	case '>':
		ts.index++
		if !ts.reachEnd() && ts.ch() == '=' {
			ts.index++
			ts.setCurrent(TOKEN_OP_GE, begin)
		} else {
			ts.setCurrent(TOKEN_OP_GT, begin)
		}
		return true
	case 'A', 'a':
		ts.index++
		if ch := ts.ch(); ch != 'N' && ch != 'n' {
			return ts.nextID(begin)
		}
		ts.index++
		if ch := ts.ch(); ch != 'D' && ch != 'd' {
			return ts.nextID(begin)
		}
		ts.index++
		if ch := ts.ch(); ts.isIdChars(ch, false) {
			return ts.nextID(begin)
		}
		ts.setCurrent(TOKEN_AND, begin)
		return true
	case 'O', 'o':
		ts.index++
		if ch := ts.ch(); ch != 'R' && ch != 'r' {
			return ts.nextID(begin)
		}
		ts.index++
		if ch := ts.ch(); ts.isIdChars(ch, false) {
			return ts.nextID(begin)
		}
		ts.setCurrent(TOKEN_OR, begin)
		return true
	case 'I', 'i':
		ts.index++
		if ch := ts.ch(); ch != 'N' && ch != 'n' {
			return ts.nextID(begin)
		}
		ts.index++
		if ch := ts.ch(); ts.isIdChars(ch, false) {
			return ts.nextID(begin)
		}
		ts.setCurrent(TOKEN_OP_IN, begin)
		return true
	case 'N', 'n':
		ts.index++
		if ch := ts.ch(); ch != 'O' && ch != 'o' {
			return ts.nextID(begin)
		}
		ts.index++
		if ch := ts.ch(); ch != 'T' && ch != 't' {
			return ts.nextID(begin)
		}
		ts.index++
		if ch := ts.ch(); ts.isIdChars(ch, false) {
			return ts.nextID(begin)
		}
		ts.setCurrent(TOKEN_NOT, begin)
		return true
	}
	return false
}

func (ts *TokenStream) nextID(begin int) bool {
	for ts.isIdChars(ts.ch(), false) {
		ts.index++
	}
	ts.setCurrent(TOKEN_ID, begin)
	return true
}

func (ts *TokenStream) isIdChars(ch rune, canBegin bool) bool {
	if ch >= 'A' && ch <= 'Z' {
		return true
	}
	if ch >= 'a' && ch <= 'z' {
		return true
	}
	if ch == '_' || ch == '$' {
		return true
	}
	if !canBegin && ch >= '0' && ch <= '9' {
		return true
	}
	return false
}

func (ts *TokenStream) nextInt(begin int) bool {
	for {
		ts.index++
		ch := ts.ch()
		if ch < '0' || ch > '9' {
			break
		}
	}
	ts.setCurrent(TOKEN_INT, begin)
	return true
}

func (ts *TokenStream) nextStr(begin int) bool {
	escaping := false
	for !ts.reachEnd() {
		ts.index++
		ch := ts.ch()
		if !escaping {
			if ch == '\\' {
				escaping = true
			} else if ch == '\'' {
				break
			}
		} else {
			escaping = false
		}
	}
	ts.index++
	ts.setCurrent(TOKEN_STR, begin)
	return true
}

func (ts *TokenStream) setCurrent(typ int, begin int) {
	ts.Current.Type = typ
	ts.Current.Text = ts.chars[begin:ts.index]
	ts.Current.Offset = begin
}

func (ts *TokenStream) reachEnd() bool {
	return ts.index >= len(ts.chars)
}

func (ts *TokenStream) skipSpaces() {
	for !ts.reachEnd() && unicode.IsSpace(ts.chars[ts.index]) {
		ts.index++
	}
}

func tokenToStr(text []rune) string {
	text = text[1 : len(text)-1] // 去掉引号
	var b strings.Builder
	escaping := false
	for _, ch := range text {
		if !escaping {
			if ch == '\\' {
				escaping = true
			} else {
				b.WriteRune(ch)
			}
		} else {
			switch ch {
			case 't':
				b.WriteRune('\t')
			case 'n':
				b.WriteRune('\n')
			case 'r':
				b.WriteRune('\r')
			case '\'':
				b.WriteRune('\'')
			default:
				b.WriteRune(ch)
			}
			escaping = false
		}
	}
	return b.String()
}

func tokenToInt(text []rune) int {
	n := len(text)
	if n < 1 {
		panic("invalid int literal")
	}
	negative := false
	i := 0
	if text[0] == '-' {
		negative = true
		i++
	} else if text[0] == '+' {
		i++
	}
	r := 0
	for i < n {
		ch := text[i]
		i++
		r = r*10 + (int(ch) - int('0'))
	}
	if negative {
		return -r
	}
	return r
}
