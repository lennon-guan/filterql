package filterql

import "fmt"

type ParseError struct {
	Err error
	Pos int
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("%d: %+v", e.Pos, e.Err)
}

func parseError(err error, pos int) *ParseError {
	return &ParseError{Err: err, Pos: pos}
}

func Parse(code string, cfg *ParseConfig) (BoolAst, error) {
	if cfg == nil {
		cfg = &defaultConfig
	}
	if cacher := cfg.Cache; cacher != nil {
		if cond, found := cacher.Load(code); found {
			return cond, nil
		}
	}
	ts := NewTokenStream(code)
	ts.Next()
	if cond, err := parseCondition(ts, cfg); err != nil {
		return nil, err
	} else {
		if cacher := cfg.Cache; cacher != nil {
			cacher.Store(code, cond)
		}
		return cond, nil
	}
}

func parseCondition(ts *TokenStream, cfg *ParseConfig) (BoolAst, error) {
	children := make([]BoolAst, 1)
	item, err := parseItem(ts, cfg)
	if err != nil {
		return nil, err
	}
	children[0] = item
	for ts.Current.Type == TOKEN_OR {
		if !ts.Next() {
			return nil, parseError(ErrUnexpectedEnd, ts.index)
		}
		item, err := parseItem(ts, cfg)
		if err != nil {
			return nil, err
		}
		children = append(children, item)
	}
	if len(children) > 1 {
		return &ORs{Children: children}, nil
	} else {
		return children[0], nil
	}
}

func parseItem(ts *TokenStream, cfg *ParseConfig) (BoolAst, error) {
	var children []BoolAst
	atom, err := parseAtom(ts, cfg)
	if err != nil {
		return nil, err
	}
	children = []BoolAst{atom}
	for ts.Current.Type == TOKEN_AND {
		if !ts.Next() {
			return nil, parseError(ErrUnexpectedEnd, ts.index)
		}
		atom, err := parseAtom(ts, cfg)
		if err != nil {
			return nil, err
		}
		children = append(children, atom)
	}
	if len(children) > 1 {
		return &ANDs{Children: children}, nil
	} else {
		return children[0], nil
	}
}

func parseAtom(ts *TokenStream, cfg *ParseConfig) (BoolAst, error) {
	if ts.Current.Type == TOKEN_LEFT_BRACKET {
		if !ts.Next() {
			return nil, parseError(ErrUnexpectedEnd, ts.index)
		}
		if cond, err := parseCondition(ts, cfg); err != nil {
			return nil, err
		} else if ts.Current.Type != TOKEN_RIGHT_BRACKET {
			return nil, parseError(ErrUnexpectedToken, ts.index)
		} else {
			ts.Next()
			return cond, nil
		}
	} else if ts.Current.Type == TOKEN_NOT {
		if !ts.Next() {
			return nil, parseError(ErrUnexpectedEnd, ts.index)
		}
		atom, err := parseAtom(ts, cfg)
		if err != nil {
			return nil, err
		} else if n, is := atom.(CanNot); is {
			return n.Not(), nil
		} else {
			return &NOT{Child: atom}, nil
		}
	}
	call, err := parseCall(ts, cfg)
	if err != nil {
		return nil, err
	}
	op := ts.Current.Type
	switch op {
	case TOKEN_OP_EQ, TOKEN_OP_NE, TOKEN_OP_GT, TOKEN_OP_GE, TOKEN_OP_LT, TOKEN_OP_LE:
		if typ, err := nextMustBe(ts, TOKEN_STR, TOKEN_INT, TOKEN_ID); err != nil {
			return nil, err
		} else {
			if typ == TOKEN_INT {
				defer ts.Next()
				return newCallThenCompare(call, op, tokenToInt(ts.Current.Text)), nil
			} else if typ == TOKEN_STR {
				defer ts.Next()
				return newCallThenCompare(call, op, tokenToStr(ts.Current.Text)), nil
			} else if call2, err := parseCall(ts, cfg); err != nil {
				return nil, err
			} else {
				return &CompareWithCall{Left: call, Op: op, Right: call2}, nil
			}
		}
		return nil, parseError(ErrUnexpectedToken, ts.index)
	case TOKEN_OP_IN:
		if typ, err := nextMustBe(ts, TOKEN_LEFT_BRACKET, TOKEN_ID); err != nil {
			return nil, err
		} else if typ == TOKEN_ID {
			if call2, err := parseCall(ts, cfg); err != nil {
				return nil, err
			} else {
				return &InWithCall{Left: call, Right: call2}, nil
			}
		}
		choiceType, err := nextMustBe(ts, TOKEN_INT, TOKEN_STR)
		if err != nil {
			return nil, err
		}
		choices := [][]rune{ts.Current.Text}
		for {
			spType, err := nextMustBe(ts, TOKEN_COMMA, TOKEN_RIGHT_BRACKET)
			if err != nil {
				return nil, err
			}
			if spType == TOKEN_RIGHT_BRACKET {
				break
			}
			if _, err := nextMustBe(ts, choiceType); err != nil {
				return nil, err
			}
			choices = append(choices, ts.Current.Text)
		}
		ts.Next()
		switch choiceType {
		case TOKEN_INT:
			if len(choices) == 1 {
				return newCallThenCompare(call, TOKEN_OP_EQ, tokenToInt(choices[0])), nil
			}
			in := &In[int]{Call: call, Choices: make([]int, len(choices))}
			for i, choice := range choices {
				in.Choices[i] = tokenToInt(choice)
			}
			return newCallThenIn(in.Call, in.Choices), nil
		case TOKEN_STR:
			if len(choices) == 1 {
				return newCallThenCompare(call, TOKEN_OP_EQ, tokenToStr(choices[0])), nil
			}
			in := &In[string]{Call: call, Choices: make([]string, len(choices))}
			for i, choice := range choices {
				in.Choices[i] = tokenToStr(choice)
			}
			return newCallThenIn(in.Call, in.Choices), nil
		default:
			panic("invalid choice type")
		}
	default:
		return call, nil
	}
}

func parseCall(ts *TokenStream, cfg *ParseConfig) (Call, error) {
	if ts.Current.Type != TOKEN_ID {
		return nil, parseError(ErrUnexpectedToken, ts.index)
	}
	name := string(ts.Current.Text)
	if _, err := nextMustBe(ts, TOKEN_LEFT_BRACKET); err != nil {
		return nil, err
	}
	var call Call
	if typ, err := nextMustBe(ts, TOKEN_STR, TOKEN_INT); err != nil {
		return nil, err
	} else {
		switch typ {
		case TOKEN_INT:
			call, err = newCall(cfg.IntMethods, name, tokenToInt(ts.Current.Text))
		case TOKEN_STR:
			call, err = newCall(cfg.StrMethods, name, tokenToStr(ts.Current.Text))
		}
		if err != nil {
			return nil, err
		}
	}
	if _, err := nextMustBe(ts, TOKEN_RIGHT_BRACKET); err != nil {
		return nil, err
	}
	ts.Next()
	return call, nil
}

func nextMustBe(ts *TokenStream, types ...int) (int, error) {
	if !ts.Next() {
		return TOKEN_NONE, parseError(ErrUnexpectedEnd, ts.index)
	}
	for _, t := range types {
		if ts.Current.Type == t {
			return t, nil
		}
	}
	return TOKEN_NONE, parseError(ErrUnexpectedToken, ts.index)
}
