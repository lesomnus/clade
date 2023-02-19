package builder

type BuilderFactory func(conf BuilderConfig) (Builder, error)

var builder_registry = map[string]BuilderFactory{
	"docker-cmd": func(conf BuilderConfig) (Builder, error) { return NewDockerCmdBuilder(conf) },
}

func Register(name string, factory BuilderFactory) {
	builder_registry[name] = factory
}
