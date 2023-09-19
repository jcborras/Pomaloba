package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"time"
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
	fmt.Print("Not implemented yet\n")
	os.Exit(22)
	return ""
}

func notSupported(cfg Configuration) string {
	strOut := "Input option combination not supported.\n"
	strOut += "Most likely you used the '-delete' option "
	strOut += "with some unsupported output type\n"
	fmt.Print(strOut)
	os.Exit(22)  // EINVAL
	return ""
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

func IpTablesRulesCommon(cfg Configuration) string {
	var dnatChain = fmt.Sprintf("POMALOBA_%s_DNAT", cfg.App)
	outStr := fmt.Sprintf(":%s - [0:0]\n", dnatChain)
	var masqChain = fmt.Sprintf("POMALOBA_%s_MASQ", cfg.App)
	outStr += fmt.Sprintf(":%s - [0:0]\n", masqChain)
	return outStr
}

func IpTablesRulesForPreroutingTable(cfg Configuration) string {
	var dnatChain = fmt.Sprintf("POMALOBA_%s_DNAT", cfg.App)
	outStr := fmt.Sprintf("-A PREROUTING -j %s\n", dnatChain)
	tplt_pre := "-A " + dnatChain + " -d %s/32 -p tcp -m tcp" + " --dport %d" +
		" -m statistic --mode random --probability %.5f" +
		" -j DNAT --to-destination %s:%d\n"
	for _, ep := range cfg.Endpoints {
		for j, dst := range cfg.Destinations {
			p := 1.0 / float32(len(cfg.Destinations)-j)
			outStr += fmt.Sprintf(tplt_pre, ep.IP, ep.Port, p, dst.IP, dst.Port)
		}
	}
	return outStr
}

func IpTablesRulesForPostroutingTable(cfg Configuration) string {
	var masqChain = fmt.Sprintf("POMALOBA_%s_MASQ", cfg.App)
	outStr := fmt.Sprintf("-A POSTROUTING -j %s\n", masqChain)
	tplt_post := "-A " + masqChain + " -d %s/32 -p tcp -m tcp --dport %d -j MASQUERADE\n" 
	for _, dst := range cfg.Destinations {
		outStr += fmt.Sprintf(tplt_post, dst.IP, dst.Port)
	}
	return outStr
}

func IpTablesRulesFor(cfg Configuration) string {
	return IpTablesRulesCommon(cfg) +
		IpTablesRulesForPreroutingTable(cfg) +
	 	IpTablesRulesForPostroutingTable(cfg)
}

func IpTablesRulesOutputFor(cfg Configuration) string {
	// TODO man 8 iptables.save
	outStr := "# Generated by Pomaloba v0.0.1 on " +
		time.Now().Format(time.RFC3339) + "\n" +
		"*nat\n" + 
		 IpTablesRulesFor(cfg) +
		"COMMIT\n" +
		"# Completed on "  + time.Now().Format(time.RFC3339) + "\n"
	return outStr
}

// Returns an Ansible playbook that realizes a LB specification
func AnsibleOutputFor(cfg Configuration) string {
	// Create new DNAT custom chain in table nat
	// https://docs.ansible.com/ansible/latest/collections/ansible/builtin/iptables_module.html
	// chain_management
	// Add rules to new DNAT chain
	// Append chain to builtin PREROUTING chain
	// Create new MASQ custom chain in table nat
	// Add rules to new MASQ chain
	// Append chain to builtin POSTROUTING chain
	
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
	"iptables-rules": notSupported,
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

func funcTablesFor(delete bool) map[string](func(Configuration) string) {
	if delete {
		return del_funcs
	}
	return gen_funcs
}

// Returns the function that generates the output from the output type
func ChooseOutputForm(form string, delete bool) func(Configuration) string {
	f, found := funcTablesFor(delete)[form]
	if !found {
		panic(fmt.Sprintf("Key '%s' for delete=%t not found", form, delete))
	}
	return f
}

// Parse the command line for options and argumets the return the
// spec file and the output format type
func getWorkToDoFromCmdline() (*string, *string, *bool) {
	help := flag.Bool("help", false, "Show help")

	deleteHelpStr := "Generate deletion rules to undo things.\n"
	deleteHelpStr += "Only supported for the 'iptables' output form but it can be\n"
	deleteHelpStr += "used to undo changes the done with the 'iptables-rules' form"
	deleteOps := flag.Bool("delete", false, deleteHelpStr)

	inSpecHelpStr := "Load Balancer specification (in JSON format)"
	inSpec := flag.String("input-spec", "", inSpecHelpStr)

	outputTypeHelpStr := "Format of the result. Accepted values are:\n"
	outputTypeHelpStr += "iptables | iptables-ansible | iptables-rules | "
	outputTypeHelpStr += "nftables | nftables-ansible | nftables-ansible"
	outputType := flag.String("output-type", "iptables", outputTypeHelpStr)


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
	f := ChooseOutputForm(*outputType, *deleteOps)
	fmt.Print(f(cfg))
}

/* Local Variables: */
/* mode: Go */
/* tab-width: 3 */
/* End: */
