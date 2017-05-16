package application

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

// Config holds the configuration
type Config struct {
	Version  string
	Bind     string
	Storages map[string]*Storage `yaml:"storages"`
	Origins  []string            `yaml:"origins"`
}

// NewConfig parses the given data and returns a representative Config struct
func NewConfig(data []byte) (config *Config, err error) {

	config = new(Config)
	if err = yaml.Unmarshal(data, &config); err != nil {
		err = fmt.Errorf("Error parsing configuration file: %v", err)
		return
	}

	fmt.Println(config)

	for name, storage := range config.Storages {
		fmt.Println(name, storage)
	}

	return
}
