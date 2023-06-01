package filterql

type ParseConfig struct {
	StrMethods map[string]func(any, string) (any, error)
	IntMethods map[string]func(any, int) (any, error)
	Cache      CacheProvider
}

var defaultConfig = ParseConfig{
	StrMethods: map[string]func(any, string) (any, error){},
	IntMethods: map[string]func(any, int) (any, error){},
}
