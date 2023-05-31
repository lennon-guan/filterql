package filterql

type Context struct {
	Env    any
	result any
}

func NewContext(env any) *Context {
	return &Context{Env: env}
}
