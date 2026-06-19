package config

type GreetConfig struct {
	// Greeting format string; %s is replaced with the name argument.
	Format string `yaml:"format"`
}
