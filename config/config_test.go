package config

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
	filesystem afero.Fs
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func (c *ConfigTestSuite) SetupTest() {
	appFS := afero.NewOsFs()
	c.filesystem = appFS
	c.filesystem.MkdirAll("etc/docker-firewall", 0755)
}

func (c *ConfigTestSuite) TearDownSuite() {
	c.filesystem.RemoveAll("etc")
}
func (c *ConfigTestSuite) Test_Config_Success() {
	configExpected := &Configuration{}
	rule1 := Rule{
		Interface: []string{"eth0"},
		Protocol:  "tcp",
		Port:      3000,
		Allow:     []string{"10.1.1.1", "10.2.1.2", "172.18.9.5", "192.168.1.15"},
	}

	rule2 := Rule{
		Port:     6000,
		Protocol: "tcp",
		Allow:    []string{"10.1.1.1", "10.2.1.2", "172.18.9.5", "192.168.1.15"},
	}

	rule3 := Rule{
		Port: 8080,
	}

	configExpected.Config.Rules = append(configExpected.Config.Rules, rule1, rule2, rule3)

	var configYaml = []byte(`
config:
  rules:
  - interface:
    - eth0
    protocol: tcp
    port: 3000
    allow:
    - 10.1.1.1
    - 10.2.1.2
    - 172.18.9.5
    - 192.168.1.15
  - port: 6000
    allow:
    - 10.1.1.1
    - 10.2.1.2
    - 172.18.9.5
    - 192.168.1.15
    protocol: tcp
  - port: 8080
`)

	afero.WriteFile(c.filesystem, "etc/docker-firewall/config.yml", configYaml, 0644)

	config, err := NewConfiguration("etc/docker-firewall")

	c.NoError(err)

	c.IsType(&Configuration{}, config)
	c.Equal(configExpected, config)

	c.filesystem.Remove("config.yml")
}

func (c *ConfigTestSuite) Test_Config_FileNotFound() {
	_, err := NewConfiguration("etc1/docker-firewall")
	c.EqualError(err, "etc1/docker-firewall/config.yml did not exist: stat etc1/docker-firewall/config.yml: no such file or directory")
}

func (c *ConfigTestSuite) Test_Config_UnableToDecode() {
	var configYaml = []byte(`
config:
   :rules:
test::alberto
`)

	afero.WriteFile(c.filesystem, "etc/docker-firewall/config.yml", configYaml, 0644)
	_, err := NewConfiguration("etc/docker-firewall")
	c.Errorf(err, "configuration error: While parsing config: yaml: line 5: could not find expected ':'")
}
