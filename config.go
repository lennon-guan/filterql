package filterql

type ParseConfig struct {
	StrMethods       map[string]func(any, string) (any, error)
	IntMethods       map[string]func(any, int) (any, error)
	Cache            CacheProvider
	DefaultIntMethod func(string, any, int) (any, error)
	DefaultStrMethod func(string, any, string) (any, error)
}

var defaultConfig = ParseConfig{
	StrMethods: map[string]func(any, string) (any, error){},
	IntMethods: map[string]func(any, int) (any, error){},
}
