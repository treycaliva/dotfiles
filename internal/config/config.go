package config

import (
	_ "embed"
	"fmt"

	"gopkg.in/yaml.v3"
)

//go:embed config.yaml
var configData []byte

type Package struct {
	Description string            `yaml:"description"`
	Deps        []string          `yaml:"deps"`
	PkgNames    map[string]string `yaml:"pkg_names"`
	Validate    string            `yaml:"validate"`
}

type Profile struct {
	Description string   `yaml:"description"`
	Packages    []string `yaml:"packages"`
}

type Config struct {
	Packages map[string]Package `yaml:"packages"`
	Profiles map[string]Profile `yaml:"profiles"`
	order    []string
}

type rawConfig struct {
	Packages yaml.Node `yaml:"packages"`
	Profiles yaml.Node `yaml:"profiles"`
}

func Load() (*Config, error) {
	var raw rawConfig
	if err := yaml.Unmarshal(configData, &raw); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	var order []string
	if raw.Packages.Kind == yaml.MappingNode {
		for i := 0; i < len(raw.Packages.Content)-1; i += 2 {
			order = append(order, raw.Packages.Content[i].Value)
		}
	}

	var cfg Config
	if err := yaml.Unmarshal(configData, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	cfg.order = order

	return &cfg, nil
}

func (c *Config) PackageNames() []string {
	return c.order
}
