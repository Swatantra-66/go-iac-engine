package parser

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Resource struct {
	Type         string `yaml:"type"`
	Name         string `yaml:"name"`
	Region       string `yaml:"region"`
	AMI          string `yaml:"ami,omitempty"`
	InstanceType string `yaml:"instance_type,omitempty"`
}

type Config struct {
	Resources []Resource `yaml:"resources"`
}

// ParseConfig reads the YAML file and unmarshals it into our structs
func ParseConfig(filepath string) (*Config, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &config, nil
}
