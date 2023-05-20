package config

import "fmt"

type Config struct{}

func New() *Config {
	return &Config{}
}

func (l *Config) CurrentPlayer() string {
	return "ZenAviator"
}

func (l *Config) GamePoint(path string) string {
	return fmt.Sprintf("http://localhost:8111/%s", path)
}
