package config

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v3"
)

type RateLimitConfig struct {
	Capacity int `yaml:"capacity"`
	Rate     int `yaml:"rate"`
}

type HealthCheckConfig struct {
	Interval string `yaml:"interval"`
	Path     string `yaml:"path"`
}

type Config struct {
	Backends    []string          `yaml:"backends"`
	Port        string            `yaml:"port"`
	RateLimit   RateLimitConfig   `yaml:"ratelimit"`
	HealthCheck HealthCheckConfig `yaml:"healthcheck"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
