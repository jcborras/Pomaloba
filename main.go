package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

// A pair (IP address, port number)
type EndPoint struct {
	IP   string
	Port uint16
}

// The transformation of a LB specification into a Go type
type Configuration struct {
	Pomaloba     int        `json:"pomaloba"`
	App          string     `json:"app"`
	Endpoints    []EndPoint `json:"endpoints"`
	Destinations []EndPoint `json:"destinations"`
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func notImplementedYet(cfg Configuration) string {
	return "Not implemented yet\n"
}

func IpTablesDeletionsForPreroutingTable(cfg Configuration) string {
	var dnatChain = fmt.Sprintf("POMALOBA_%s_DNAT", cfg.App)
	outStr := fmt.Sprintf("iptables --table nat --delete PREROUTING --jump %s\n", dnatChain)
	tplt_pre := "iptables --table nat --delete " + dnatChain + " %d\n"
	for i, _ := range cfg.Endpoints {
		for j, _ := range cfg.Destinations {
			k := len(cfg.Endpoints)*len(cfg.Destinations)-(i*len(cfg.Endpoints)+j)
			outStr += fmt.Sprintf(tplt_pre, k)
		}
	}
	outStr += fmt.Sprintf("iptables --table nat --delete-chain %s\n", dnatChain)
	return outStr
}

func IpTablesDeletionsForPostroutingTable(cfg Configuration) string {
	var masqChain = fmt.Sprintf("POMALOBA_%s_MASQ", cfg.App)
	outStr := fmt.Sprintf("iptables --table nat --delete POSTROUTING --jump %s\n", masqChain)
	tplt_post := "iptables --table nat --delete " + masqChain +  " %d\n"
	for i, _ := range cfg.Destinations {
		outStr += fmt.Sprintf(tplt_post, len(cfg.Destinations)-i)
	}
	outStr += fmt.Sprintf("iptables --table nat --delete-chain %s\n", masqChain)
	return outStr
}

func IpTablesDeletionstFor(cfg Configuration) string {
	return IpTablesDeletionsForPreroutingTable(cfg) +
	       IpTablesDeletionsForPostroutingTable(cfg)
}

func IpTablesOutputForPreroutingTable(cfg Configuration) string {
	var dnatChain = fmt.Sprintf("POMALOBA_%s_DNAT", cfg.App)
	outStr := fmt.Sprintf("iptables --table nat --new-chain %s\n", dnatChain)

	tplt_pre := "iptables" +
		" --table nat --append " + dnatChain + " " +
		" --protocol tcp --destination %s --dport %d" +
		" --match statistic --mode random --probability %.5f" +
		" --jump DNAT --to-destination %s:%d\n"
	for _, ep := range cfg.Endpoints {
		for j, dst := range cfg.Destinations {
			p := 1.0 / float32(len(cfg.Destinations)-j)
			outStr += fmt.Sprintf(tplt_pre, ep.IP, ep.Port, p, dst.IP, dst.Port)
		}
	}
	outStr += fmt.Sprintf("iptables --table nat --append PREROUTING --jump %s\n",
		dnatChain)
	return outStr
}

func IpTablesOutputForPostroutingTable(cfg Configuration) string {
	var masqChain = fmt.Sprintf("POMALOBA_%s_MASQ", cfg.App)
	outStr := fmt.Sprintf("iptables --table nat --new-chain %s\n", masqChain)

	tplt_post := "iptables" +
		" --table nat --append " + masqChain +  
		" --protocol tcp --destination %s --dport %d" +
		" --jump MASQUERADE\n"
	for _, dst := range cfg.Destinations {
		outStr += fmt.Sprintf(tplt_post, dst.IP, dst.Port)
	}
	outStr += fmt.Sprintf("iptables --table nat --append POSTROUTING --jump %s\n",
		masqChain)
	return outStr
}

// Returns the iptables statement that realize a Load Balancing specification
func IpTablesOutputFor(cfg Configuration) string {
	return IpTablesOutputForPreroutingTable(cfg) +
	       IpTablesOutputForPostroutingTable(cfg)
}

func IpTablesRulesOutputFor(cfg Configuration) string {
	// TODO man 8 iptables.save
	return "# Generated by Pomaloba v0.1 at date\n" +
		"TODO\n" + "COMMIT\n" +
		"# Completed on date\n"
}

// Returns an Ansible playbook that realizes a LB specification
func AnsibleOutputFor(cfg Configuration) string {
	tplt := "- name: This is a test, there are other tests like it, but..\n" +
		"  ansible.builtin.iptables:\n" +
		"  chain: FOO\n" +
		"  table: nat\n" +
		"  protocol: tcp\n"
	return tplt
}

var del_funcs = map[string](func(Configuration) string){
	"ansible":  notImplementedYet,
	"iptables": IpTablesDeletionstFor,
	"iptables-rules": notImplementedYet,
	"nftables": notImplementedYet,
	"nftables-rules": notImplementedYet,
}

var gen_funcs = map[string](func(Configuration) string){
	"ansible": AnsibleOutputFor,
	"iptables": IpTablesOutputFor,
	"iptables-rules": IpTablesRulesOutputFor,
	"nftables": notImplementedYet,
	"nftables-rules": notImplementedYet,
}

// Returns the function that generates the output from the output type
func ChooseOutputForm(form string, delete bool) func(Configuration) string {
	var m map[string](func(Configuration) string)
	if delete {
		m = del_funcs
	} else {
		m = gen_funcs
	}
	if f, found := m[form]; !found {
		panic(fmt.Sprintf("Key '%s' for delete=%b not found", form, delete))
	} else {
		return f
	}
	
}

// Parse the command line for options and argumets the return the
// spec file and the output format type
func getWorkToDoFromCmdline() (*string, *string, *bool) {
	help := flag.Bool("help", false, "Show help")
	deleteOps := flag.Bool("delete", false, "Generate deletion rules to undo things")
	inSpec := flag.String("input-spec", "", "Input Load Balancer spec (JSON)")
	outputType := flag.String("output-type", "iptables",
	"One of: ansible|iptables|iptables-rules|nftables|nftables-rules")

	flag.Parse()

	if *help || (len(*inSpec) == 0) {
		flag.Usage()
		os.Exit(0)
	}
	return inSpec, outputType, deleteOps
}

// Load the Load Balancing specificaton from a JSON file
func getLBSpecFromJSONFile(inSpec string) Configuration {
	dat, err1 := os.ReadFile(inSpec)
	check(err1)

	var cfg Configuration
	err2 := json.Unmarshal(dat, &cfg)
	check(err2)
	return cfg
}

func main() {
	inSpec, outputType, deleteOps := getWorkToDoFromCmdline()	

	cfg := getLBSpecFromJSONFile(*inSpec)
	//fmt.Println(cfg.Pomaloba)
	//fmt.Println(cfg.App)
	//fmt.Println(cfg.Endpoints)
	//fmt.Println(cfg.Destinations)

	f := ChooseOutputForm(*outputType, *deleteOps)
	fmt.Print(f(cfg))
}

/* Local Variables: */
/* mode: Go */
/* tab-width: 3 */
/* End: */
