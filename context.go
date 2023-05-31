package filterql

type Context struct {
	result     any
	Env        any
	intMethods map[string]func(any, int) (any, error)
	strMethods map[string]func(any, string) (any, error)
}

func NewContext(env any) *Context {
	return NewContextWithMethods(env, nil, nil)
}

func NewContextWithMethods(
	env any,
	ims map[string]func(any, int) (any, error),
	sms map[string]func(any, string) (any, error),
) *Context {
	if ims == nil {
		ims = map[string]func(any, int) (any, error){}
	}
	if sms == nil {
		sms = map[string]func(any, string) (any, error){}
	}
	return &Context{
		Env:        env,
		intMethods: ims,
		strMethods: sms,
	}
}

func (ctx *Context) invokeInt(name string, param int) (any, error) {
	if f, found := ctx.intMethods[name]; !found {
		return nil, ErrNoSuchMethod
	} else {
		return f(ctx.Env, param)
	}
}

func (ctx *Context) invokeStr(name string, param string) (any, error) {
	if f, found := ctx.strMethods[name]; !found {
		return nil, ErrNoSuchMethod
	} else {
		return f(ctx.Env, param)
	}
}
