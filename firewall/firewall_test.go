package firewall

import (
	"fmt"
	"testing"

	"github.com/albertogviana/docker-firewall/config"
	"github.com/coreos/go-iptables/iptables"
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

func (f *FirewallTestSuite) Test_Rules() {
	configuration := &config.Configuration{}
	rule1 := config.Rule{
		Interface: []string{"eth0"},
		Protocol:  "tcp",
		Port:      8080,
		Allow:     []string{"10.1.1.1"},
	}

	rule2 := config.Rule{
		Port:  8080,
		Allow: []string{"10.1.1.1", "10.2.1.2", "172.18.9.5", "192.168.1.15"},
	}

	rule3 := config.Rule{
		Port: 8080,
	}

	rule4 := config.Rule{
		Interface: []string{"docker_gwbridge"},
	}

	configuration.Config.Rules = append(configuration.Config.Rules, rule1, rule2, rule3, rule4)

	firewall, err := NewFirewall()
	f.NoError(err)

	err = firewall.Apply(configuration.Config.Rules)
	f.NoError(err)

	expectedRules := [][]string{
		{"-j", "DROP"},
		{"-m", "conntrack", "--ctstate", "RELATED,ESTABLISHED", "-j", "RETURN"},
		{"-s", "10.1.1.1", "-i", "eth0", "-p", "tcp", "-m", "tcp", "--dport", "8080", "-j", "RETURN"},
		{"-s", "10.1.1.1", "-p", "tcp", "-m", "tcp", "--dport", "8080", "-j", "RETURN"},
		{"-s", "10.1.1.1", "-p", "udp", "-m", "udp", "--dport", "8080", "-j", "RETURN"},
		{"-s", "10.2.1.2", "-p", "tcp", "-m", "tcp", "--dport", "8080", "-j", "RETURN"},
		{"-s", "10.2.1.2", "-p", "udp", "-m", "udp", "--dport", "8080", "-j", "RETURN"},
		{"-s", "172.18.9.5", "-p", "tcp", "-m", "tcp", "--dport", "8080", "-j", "RETURN"},
		{"-s", "172.18.9.5", "-p", "udp", "-m", "udp", "--dport", "8080", "-j", "RETURN"},
		{"-s", "192.168.1.15", "-p", "tcp", "-m", "tcp", "--dport", "8080", "-j", "RETURN"},
		{"-s", "192.168.1.15", "-p", "udp", "-m", "udp", "--dport", "8080", "-j", "RETURN"},
		{"-p", "tcp", "-m", "tcp", "--dport", "8080", "-j", "RETURN"},
		{"-p", "udp", "-m", "udp", "--dport", "8080", "-j", "RETURN"},
		{"-i", "docker_gwbridge", "-j", "RETURN"},
	}

	ipt, err := iptables.New()
	f.NoError(err)

	for _, rule := range expectedRules {
		exists, err := ipt.Exists(FilterTable, DockerUserChain, rule...)
		f.NoError(err)

		var msg interface{}
		msg = fmt.Sprintf("Rule %s not found", rule)
		f.True(exists, msg)
	}

	verifyRules, err := firewall.Verify(configuration.Config.Rules)
	f.NoError(err)
	f.True(verifyRules)

	// firewall.ClearRule()
	// for _, rule := range expectedRules {
	// 	exists, err := ipt.Exists(FilterTable, DockerUserChain, rule...)
	// 	f.NoError(err)

	// 	var msg interface{}
	// 	msg = fmt.Sprintf("Rule %s not found", rule)
	// 	f.False(exists, msg)
	// }
}

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
				{"-s", "10.1.1.1", "-i", "eth0", "-p", "tcp", "-m", "tcp", "--dport", "8080", "-j", "RETURN"},
				{"-s", "10.1.1.1", "-i", "eth1", "-p", "tcp", "-m", "tcp", "--dport", "8080", "-j", "RETURN"},
				{"-s", "192.168.10.11", "-i", "eth0", "-p", "tcp", "-m", "tcp", "--dport", "8080", "-j", "RETURN"},
				{"-s", "192.168.10.11", "-i", "eth1", "-p", "tcp", "-m", "tcp", "--dport", "8080", "-j", "RETURN"},
			},
		},
		{
			config.Rule{
				Port:  8080,
				Allow: []string{"10.1.1.1", "10.2.1.2", "172.18.9.5", "192.168.1.15"},
			},
			[][]string{
				{"-s", "10.1.1.1", "-p", "tcp", "-m", "tcp", "--dport", "8080", "-j", "RETURN"},
				{"-s", "10.1.1.1", "-p", "udp", "-m", "udp", "--dport", "8080", "-j", "RETURN"},
				{"-s", "10.2.1.2", "-p", "tcp", "-m", "tcp", "--dport", "8080", "-j", "RETURN"},
				{"-s", "10.2.1.2", "-p", "udp", "-m", "udp", "--dport", "8080", "-j", "RETURN"},
				{"-s", "172.18.9.5", "-p", "tcp", "-m", "tcp", "--dport", "8080", "-j", "RETURN"},
				{"-s", "172.18.9.5", "-p", "udp", "-m", "udp", "--dport", "8080", "-j", "RETURN"},
				{"-s", "192.168.1.15", "-p", "tcp", "-m", "tcp", "--dport", "8080", "-j", "RETURN"},
				{"-s", "192.168.1.15", "-p", "udp", "-m", "udp", "--dport", "8080", "-j", "RETURN"},
			},
		},
		{
			config.Rule{
				Port: 8080,
			},
			[][]string{
				{"-p", "tcp", "-m", "tcp", "--dport", "8080", "-j", "RETURN"},
				{"-p", "udp", "-m", "udp", "--dport", "8080", "-j", "RETURN"},
			},
		},
		{
			config.Rule{
				Interface: []string{"docker_gwbridge"},
			},
			[][]string{
				{"-i", "docker_gwbridge", "-j", "RETURN"},
			},
		},
	}

	for _, test := range tests {
		f.Equal(test.expected, generateRules(test.rule))
	}
}
