package firewall

import (
	"strconv"

	"github.com/albertogviana/docker-firewall/config"
	"github.com/coreos/go-iptables/iptables"
)

// Firewall defines the firewall structure and its dependencies
type Firewall struct {
	iptables *iptables.IPTables
}

// DockerUserChain is the iptables chain used to create the rules
const DockerUserChain = "DOCKER-USER"

// FilterTable is used for packet filtering on iptables
const FilterTable = "filter"

// ReturnTarget purpose is to return from a user-defined chain before rule matching on that chain has completed.
const ReturnTarget = "RETURN"

var baseRules = [][]string{
	{"-j", "DROP"},
	{"-m", "conntrack", "--ctstate", "RELATED,ESTABLISHED", "-j", "RETURN"},
}

// NewFirewall returns a Firewall instance
func NewFirewall() (*Firewall, error) {
	firewall := &Firewall{}

	ipt, err := iptables.New()
	if err != nil {
		return nil, err
	}

	firewall.iptables = ipt

	return firewall, nil
}

// Apply parse the configuration and applying it in the system
func (f *Firewall) Apply(rules []config.Rule) error {
	iptablesRules := baseRules

	for _, rule := range rules {
		r := generateRules(rule)
		iptablesRules = append(iptablesRules, r...)
	}

	f.ClearRule()
	for _, iptRule := range iptablesRules {
		err := f.iptables.Insert(FilterTable, DockerUserChain, 1, iptRule...)
		if err != nil {
			return err
		}
	}

	return nil
}

// Verify checks if the rules in the configuration files where applied.
func (f *Firewall) Verify(rules []config.Rule) (bool, error) {
	iptablesRules := baseRules

	for _, rule := range rules {
		r := generateRules(rule)
		iptablesRules = append(iptablesRules, r...)
	}

	result := true
	for _, rule := range iptablesRules {
		exists, err := f.iptables.Exists(FilterTable, DockerUserChain, rule...)
		if err != nil {
			return false, err
		}

		if !exists {
			result = false
		}
	}

	return result, nil
}

// ClearRule cleans the DOCKER-USER chain
func (f *Firewall) ClearRule() error {
	err := f.iptables.ClearChain(FilterTable, DockerUserChain)
	if err != nil {
		return err
	}

	err = f.iptables.Insert(FilterTable, DockerUserChain, 1, "-j", "RETURN")
	if err != nil {
		return err
	}

	return nil
}

func generateRules(rule config.Rule) [][]string {
	rules := [][]string{}

	baseRule := []string{}
	if rule.Protocol != "" {
		baseRule = append(baseRule, "-p", rule.Protocol, "-m", rule.Protocol)
	}

	if rule.Port > 0 {
		baseRule = append(baseRule, "--dport", strconv.Itoa(rule.Port))
	}

	if len(rule.Interface) > 0 {
		for _, i := range rule.Interface {
			interfaceRule := []string{}
			interfaceRule = append(interfaceRule, "-i", i)
			interfaceRule = append(interfaceRule, baseRule...)
			rules = append(rules, interfaceRule)
		}
	}

	if rule.Protocol == "" {
		tmpRules := [][]string{}
		if len(rules) == 0 {
			tcp := []string{}
			tcp = append(tcp, "-p", "tcp", "-m", "tcp")
			tcp = append(tcp, baseRule...)
			tmpRules = append(tmpRules, tcp)

			udp := []string{}
			udp = append(udp, "-p", "udp", "-m", "udp")
			udp = append(udp, baseRule...)
			tmpRules = append(tmpRules, udp)
		}
		rules = tmpRules
	}

	if len(rule.Allow) > 0 {
		tmpRules := [][]string{}
		for _, ip := range rule.Allow {
			for _, v := range rules {
				ipRule := []string{}
				ipRule = append(ipRule, "-s", ip)
				ipRule = append(ipRule, v...)
				tmpRules = append(tmpRules, ipRule)
			}
		}
		rules = tmpRules
	}

	if len(rules) > 0 {
		tmpRules := rules
		rules = [][]string{}
		for _, v := range tmpRules {
			targetRule := []string{}
			targetRule = append(targetRule, v...)
			targetRule = append(targetRule, "-j", ReturnTarget)
			rules = append(rules, targetRule)
		}
	}

	return rules
}
