package firewall

import (
	"testing"

	"github.com/coreos/go-iptables/iptables"

	"github.com/albertogviana/docker-firewall/config"
	"github.com/stretchr/testify/suite"
)

type FirewallTestSuite struct {
	suite.Suite
}

func TestFirewallTestSuite(t *testing.T) {
	suite.Run(t, new(FirewallTestSuite))
}

func (f *FirewallTestSuite) Test_NewFirewall() {
	firewall, err := NewFirewall()
	f.NoError(err)
	f.IsType(&Firewall{}, firewall)
	f.IsType(&iptables.IPTables{}, firewall.iptables)
}

// func (f *FirewallTestSuite) Test_Rules() {
// 	configuration := &config.Configuration{}
// 	rule1 := config.Rule{
// 		Interface: []string{"eth0"},
// 		Protocol:  "tcp",
// 		Port:      8080,
// 		Allow:     []string{"10.1.1.1"},
// 	}

// 	// rule2 := config.Rule{
// 	// 	Port:  8080,
// 	// 	Allow: []string{"10.1.1.1", "10.2.1.2", "172.18.9.5", "192.168.1.15"},
// 	// }

// 	// rule3 := config.Rule{
// 	// 	Port: 8080,
// 	// }

// 	// configuration.Config.Rules = append(configuration.Config.Rules, rule1, rule2, rule3)
// 	configuration.Config.Rules = append(configuration.Config.Rules, rule1)

// 	firewall, err := NewFirewall()
// 	f.NoError(err)

// 	err = firewall.Apply(configuration.Config.Rules)
// 	f.NoError(err)

// 	ipt, err := iptables.New()
// 	f.NoError(err)

// 	// err = ipt.Insert(FilterTable, DockerUserChain, 1, "-j", "DROP")
// 	// f.NoError(err)

// 	// ipt.Exists("filter", DockerUserChain, "-j", "DROP")

// 	err = ipt.ClearChain(FilterTable, DockerUserChain)
// 	f.NoError(err)
// 	err = ipt.Insert(FilterTable, DockerUserChain, 1, "-j", "RETURN")
// 	f.NoError(err)
// }

func (f *FirewallTestSuite) Test_GenerateRules() {
	var tests = []struct {
		rule     config.Rule
		expected [][]string
	}{
		{
			config.Rule{
				Interface: []string{"eth0", "eth1"},
				Protocol:  "tcp",
				Port:      8080,
				Allow:     []string{"10.1.1.1", "192.168.10.11"},
			},
			[][]string{
				{"-i", "eth0", "-p", "tcp", "-m", "tcp", "--dport", "8080", "-s", "10.1.1.1", "-j", "RETURN"},
				{"-i", "eth1", "-p", "tcp", "-m", "tcp", "--dport", "8080", "-s", "10.1.1.1", "-j", "RETURN"},
				{"-i", "eth0", "-p", "tcp", "-m", "tcp", "--dport", "8080", "-s", "192.168.10.11", "-j", "RETURN"},
				{"-i", "eth1", "-p", "tcp", "-m", "tcp", "--dport", "8080", "-s", "192.168.10.11", "-j", "RETURN"},
			},
		},
		{
			config.Rule{
				Port:  8080,
				Allow: []string{"10.1.1.1", "10.2.1.2", "172.18.9.5", "192.168.1.15"},
			},
			[][]string{
				{"--dport", "8080", "-s", "10.1.1.1", "-j", "RETURN"},
				{"--dport", "8080", "-s", "10.2.1.2", "-j", "RETURN"},
				{"--dport", "8080", "-s", "172.18.9.5", "-j", "RETURN"},
				{"--dport", "8080", "-s", "192.168.1.15", "-j", "RETURN"},
			},
		},
		{
			config.Rule{
				Port: 8080,
			},
			[][]string{
				{"--dport", "8080", "-j", "RETURN"},
			},
		},
	}

	for _, test := range tests {
		f.Equal(test.expected, generateRules(test.rule))
	}
}
