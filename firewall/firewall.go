package firewall

import (
	"fmt"
	"strconv"

	"github.com/albertogviana/docker-firewall/config"
	"github.com/coreos/go-iptables/iptables"
)

type Firewall struct {
	iptables *iptables.IPTables
}

// DockerUserChain is the iptables chain used to create the rules
const DockerUserChain = "DOCKER-USER"

// FilterTable is used for packet filtering on iptables
const FilterTable = "filter"

const ReturnTarget = "RETURN"

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
	for _, rule := range rules {
		fmt.Println(rule)
	}

	return nil
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

	if len(rule.Allow) > 0 {
		tmpRules := [][]string{}
		for _, ip := range rule.Allow {
			if len(rules) > 0 {
				for _, v := range rules {
					ipRule := []string{}
					ipRule = append(ipRule, v...)
					ipRule = append(ipRule, "-s", ip)
					tmpRules = append(tmpRules, ipRule)
				}
			}

			if len(rules) == 0 {
				ipRule := []string{}
				ipRule = append(ipRule, baseRule...)
				ipRule = append(ipRule, "-s", ip)
				tmpRules = append(tmpRules, ipRule)
			}
		}
		rules = [][]string{}
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

	if len(rules) == 0 {
		targetRule := []string{}
		targetRule = append(targetRule, baseRule...)
		targetRule = append(targetRule, "-j", ReturnTarget)
		rules = append(rules, targetRule)
	}

	return rules
}
