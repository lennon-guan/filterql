package filterql

import (
	"io"
	"reflect"
)

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

type Call struct {
	Name      string
	ParamType int
	IntParam  int
	StrParam  string
	not       bool
}

func (c *Call) Eval(ctx *Context) error {
	var err error
	if c.ParamType == TOKEN_INT {
		ctx.result, err = ctx.invokeInt(c.Name, c.IntParam)
	} else if c.ParamType == TOKEN_STR {
		ctx.result, err = ctx.invokeStr(c.Name, c.StrParam)
	} else {
		panic("invalid param type")
	}
	return err
}

func (c *Call) IsTrue(ctx *Context) (bool, error) {
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

func (c *Call) Not() BoolAst {
	return &Call{
		Name:      c.Name,
		ParamType: c.ParamType,
		IntParam:  c.IntParam,
		StrParam:  c.StrParam,
		not:       !c.not,
	}
}

func compareByOp[T int | string](a, b T, op int) bool {
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

type Compare[T int | string] struct {
	Call   *Call
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
	var op int
	switch c.Op {
	case TOKEN_OP_EQ:
		op = TOKEN_OP_NE
	case TOKEN_OP_NE:
		op = TOKEN_OP_EQ
	case TOKEN_OP_LT:
		op = TOKEN_OP_GE
	case TOKEN_OP_LE:
		op = TOKEN_OP_GT
	case TOKEN_OP_GT:
		op = TOKEN_OP_LE
	case TOKEN_OP_GE:
		op = TOKEN_OP_LT
	default:
		panic("invalid compare op")
	}
	return &Compare[T]{
		Call:   c.Call,
		Op:     op,
		Target: c.Target,
	}
}

func inSlice[T int | string](val T, slice []T) bool {
	for _, item := range slice {
		if val == item {
			return true
		}
	}
	return false
}

type In[T int | string] struct {
	Call    *Call
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
	Left, Right *Call
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

type InWithCall struct {
	Left, Right *Call
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
