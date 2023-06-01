package filterql

import (
	"io"
	"reflect"
)

type TArg interface {
	int | string
}
type PrintableAst interface {
	PrintTo(level int, out io.Writer)
}

type EvalAst interface {
	PrintableAst
	Eval(*Context) error
}

type BoolAst interface {
	PrintableAst
	IsTrue(*Context) (bool, error)
}

type CanNot interface {
	Not() BoolAst
}

type ANDs struct {
	Children []BoolAst
}

func (a *ANDs) IsTrue(ctx *Context) (bool, error) {
	for _, child := range a.Children {
		if rv, err := child.IsTrue(ctx); err != nil {
			return false, err
		} else if !rv {
			return false, nil
		}
	}
	return true, nil
}

func (a *ANDs) Not() BoolAst {
	children := make([]BoolAst, len(a.Children))
	for i, child := range a.Children {
		if n, is := child.(CanNot); is {
			children[i] = n.Not()
		} else {
			children[i] = &NOT{Child: child}
		}
	}
	return &ORs{Children: children}
}

type ORs struct {
	Children []BoolAst
}

func (a *ORs) IsTrue(ctx *Context) (bool, error) {
	for _, child := range a.Children {
		if rv, err := child.IsTrue(ctx); err != nil {
			return false, err
		} else if rv {
			return true, nil
		}
	}
	return false, nil
}

func (a *ORs) Not() BoolAst {
	children := make([]BoolAst, len(a.Children))
	for i, child := range a.Children {
		if n, is := child.(CanNot); is {
			children[i] = n.Not()
		} else {
			children[i] = &NOT{Child: child}
		}
	}
	return &ANDs{Children: children}
}

type NOT struct {
	Child BoolAst
}

func (a *NOT) IsTrue(ctx *Context) (bool, error) {
	if r, err := a.Child.IsTrue(ctx); err != nil {
		return false, err
	} else {
		return !r, nil
	}
}

func (a *NOT) Not() BoolAst {
	return a.Child
}

type Call interface {
	BoolAst
	EvalAst
	CanNot
}

type call[T TArg] struct {
	name string
	arg  T
	fn   func(any, T) (any, error)
	not  bool
}

func newCall[T TArg](fnMap map[string]func(any, T) (any, error), name string, arg T) (*call[T], error) {
	fn, has := fnMap[name]
	if !has {
		return nil, ErrNoSuchMethod
	}
	return &call[T]{
		name: name,
		arg:  arg,
		fn:   fn,
	}, nil
}

func (c *call[T]) Eval(ctx *Context) (err error) {
	ctx.result, err = c.fn(ctx.Env, c.arg)
	return
}

func (c *call[T]) IsTrue(ctx *Context) (bool, error) {
	if err := c.Eval(ctx); err != nil {
		return false, err
	}
	switch result := ctx.result.(type) {
	case int:
		return (result != 0) != c.not, nil
	case string:
		return (result != "") != c.not, nil
	case bool:
		return result != c.not, nil
	default:
		return reflect.ValueOf(result).IsZero() == c.not, nil
	}
}

func (c *call[T]) Not() BoolAst {
	return &call[T]{
		name: c.name,
		arg:  c.arg,
		fn:   c.fn,
		not:  !c.not,
	}
}

func compareByOp[T TArg](a, b T, op int) bool {
	switch op {
	case TOKEN_OP_EQ:
		return a == b
	case TOKEN_OP_NE:
		return a != b
	case TOKEN_OP_LT:
		return a < b
	case TOKEN_OP_LE:
		return a <= b
	case TOKEN_OP_GT:
		return a > b
	case TOKEN_OP_GE:
		return a >= b
	default:
		panic("invalid compare op")
	}
}

func reverseOp(op int) int {
	switch op {
	case TOKEN_OP_EQ:
		return TOKEN_OP_NE
	case TOKEN_OP_NE:
		return TOKEN_OP_EQ
	case TOKEN_OP_LT:
		return TOKEN_OP_GE
	case TOKEN_OP_LE:
		return TOKEN_OP_GT
	case TOKEN_OP_GT:
		return TOKEN_OP_LE
	case TOKEN_OP_GE:
		return TOKEN_OP_LT
	}
	panic("invalid compare op")
}

type Compare[T TArg] struct {
	Call   Call
	Op     int
	Target T
}

func (c *Compare[T]) IsTrue(ctx *Context) (bool, error) {
	if err := c.Call.Eval(ctx); err != nil {
		return false, err
	}
	if result, is := ctx.result.(T); !is {
		return false, ErrTypeNotMatched
	} else {
		return compareByOp(result, c.Target, c.Op), nil
	}
}

func (c *Compare[T]) Not() BoolAst {
	return &Compare[T]{
		Call:   c.Call,
		Op:     reverseOp(c.Op),
		Target: c.Target,
	}
}

func inSlice[T TArg](val T, slice []T) bool {
	for _, item := range slice {
		if val == item {
			return true
		}
	}
	return false
}

type In[T TArg] struct {
	Call    Call
	NotIn   bool
	Choices []T
}

func (c *In[T]) IsTrue(ctx *Context) (bool, error) {
	if err := c.Call.Eval(ctx); err != nil {
		return false, err
	}
	if result, is := ctx.result.(T); !is {
		return false, ErrTypeNotMatched
	} else {
		return inSlice(result, c.Choices) != c.NotIn, nil
	}
}

func (c *In[T]) Not() BoolAst {
	return &In[T]{
		Call:    c.Call,
		Choices: c.Choices,
		NotIn:   !c.NotIn,
	}
}

type CompareWithCall struct {
	Left, Right Call
	Op          int
}

func (c *CompareWithCall) IsTrue(ctx *Context) (bool, error) {
	if err := c.Left.Eval(ctx); err != nil {
		return false, err
	}
	res1 := ctx.result
	if err := c.Right.Eval(ctx); err != nil {
		return false, err
	}
	res2 := ctx.result
	switch v1 := res1.(type) {
	case int:
		if v2, is := res2.(int); is {
			return compareByOp(v1, v2, c.Op), nil
		}
	case string:
		if v2, is := res2.(string); is {
			return compareByOp(v1, v2, c.Op), nil
		}
	}
	return false, ErrTypeNotMatched
}

func (c *CompareWithCall) Not() BoolAst {
	return &CompareWithCall{
		Left:  c.Left,
		Right: c.Right,
		Op:    reverseOp(c.Op),
	}
}

type InWithCall struct {
	Left, Right Call
	NotIn       bool
}

func (c *InWithCall) IsTrue(ctx *Context) (bool, error) {
	if err := c.Left.Eval(ctx); err != nil {
		return false, err
	}
	res1 := ctx.result
	if err := c.Right.Eval(ctx); err != nil {
		return false, err
	}
	res2 := ctx.result
	switch v1 := res1.(type) {
	case int:
		if v2, is := res2.([]int); is {
			return inSlice(v1, v2) != c.NotIn, nil
		}
	case string:
		if v2, is := res2.([]string); is {
			return inSlice(v1, v2) != c.NotIn, nil
		}
	}
	return false, ErrTypeNotMatched
}

func (c *InWithCall) Not() BoolAst {
	return &InWithCall{
		Left:  c.Left,
		Right: c.Right,
		NotIn: !c.NotIn,
	}
}

type callThenCompare[T1, T2 TArg] struct {
	name   string
	arg    T1
	fn     func(any, T1) (any, error)
	target T2
	op     int
}

func newCallThenCompare[T TArg](ci Call, op int, target T) BoolAst {
	switch c := ci.(type) {
	case *call[int]:
		return &callThenCompare[int, T]{
			name:   c.name,
			arg:    c.arg,
			fn:     c.fn,
			target: target,
			op:     op,
		}
	case *call[string]:
		return &callThenCompare[string, T]{
			name:   c.name,
			arg:    c.arg,
			fn:     c.fn,
			target: target,
			op:     op,
		}
	}
	panic("invalid call")
}

func (c *callThenCompare[T1, T2]) IsTrue(ctx *Context) (bool, error) {
	ret, err := c.fn(ctx.Env, c.arg)
	if err != nil {
		return false, err
	} else if result, is := ret.(T2); !is {
		return false, ErrTypeNotMatched
	} else {
		return compareByOp(result, c.target, c.op), nil
	}
}

func (c *callThenCompare[T1, T2]) Not() BoolAst {
	return &callThenCompare[T1, T2]{
		name:   c.name,
		arg:    c.arg,
		fn:     c.fn,
		target: c.target,
		op:     reverseOp(c.op),
	}
}

type callThenIn[T1, T2 TArg] struct {
	name    string
	arg     T1
	fn      func(any, T1) (any, error)
	choices []T2
	not     bool
}

func newCallThenIn[T TArg](ci Call, choices []T) BoolAst {
	switch c := ci.(type) {
	case *call[int]:
		return &callThenIn[int, T]{
			name:    c.name,
			arg:     c.arg,
			fn:      c.fn,
			choices: choices,
		}
	case *call[string]:
		return &callThenIn[string, T]{
			name:    c.name,
			arg:     c.arg,
			fn:      c.fn,
			choices: choices,
		}
	}
	panic("invalid call")
}

func (c *callThenIn[T1, T2]) IsTrue(ctx *Context) (bool, error) {
	ret, err := c.fn(ctx.Env, c.arg)
	if err != nil {
		return false, err
	} else if result, is := ret.(T2); !is {
		return false, ErrTypeNotMatched
	} else {
		return inSlice(result, c.choices) != c.not, nil
	}
}

func (c *callThenIn[T1, T2]) Not() BoolAst {
	return &callThenIn[T1, T2]{
		name:    c.name,
		arg:     c.arg,
		fn:      c.fn,
		choices: c.choices,
		not:     !c.not,
	}
}
