package config

import (
	"fmt"

	"github.com/spf13/viper"
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
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath(configDirectory)
	err := viper.ReadInConfig()

	if err != nil {
		return &Configuration{}, fmt.Errorf("configuration error: %s", err)
	}

	var configuration Configuration

	err = viper.Unmarshal(&configuration)
	if err != nil {
		return &Configuration{}, fmt.Errorf("unable to decode into struct, %v", err)
	}

	return &configuration, nil
}
