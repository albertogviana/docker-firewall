package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"gopkg.in/yaml.v2"
)

// Configuration defines the configuration structure
type Configuration struct {
	Config Rules `yaml:"config"`
}

// Rules defines a list of rules
type Rules struct {
	Rules []Rule
}

// Rule defines a rule
type Rule struct {
	Interface []string `yaml:"interface,omitempty"`
	Protocol  string   `yaml:"protocol,omitempty"`
	Port      int      `yaml:"port,omitempty"`
	Allow     []string `yaml:"allow,omitempty"`
}

// NewConfiguration reads and parse the configuration file
func NewConfiguration(configDirectory string) (*Configuration, error) {
	if _, err := os.Stat(path.Join(configDirectory, "config.yml")); err != nil {
		return nil, fmt.Errorf("%s/config.yml did not exist: %v", configDirectory, err)
	}

	data, err := ioutil.ReadFile(path.Join(configDirectory, "config.yml"))
	if err != nil {
		return nil, fmt.Errorf("fail to read the file %s: %v", configDirectory, err)
	}

	var configuration Configuration

	err = yaml.Unmarshal(data, &configuration)
	if err != nil {
		return nil, fmt.Errorf("unable to decode into struct, %v", err)
	}

	return &configuration, nil
}
