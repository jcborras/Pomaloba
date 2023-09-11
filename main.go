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

// Returns the iptables statement that realize a Load Balancing specification
func IpTablesOutputFor(cfg Configuration) string {
	tplt := "iptables" +
		" --table nat --append PREROUTING" +
		" --protocol tcp --destination %s --dport %d" +
		" --match statistic --mode random --probability %.5f" +
		" --jump DNAT --to-destination %s:%d\n"
	outStr := ""
	for _, ep := range cfg.Endpoints {
		for j, dst := range cfg.Destinations {
			p := 1.0 / float32(len(cfg.Destinations)-j)
			outStr += fmt.Sprintf(tplt, ep.IP, ep.Port, p, dst.IP, dst.Port)
		}
	}
	return outStr
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

// Returns a map where the keys are valid output types and the values
// a functions that generate the actual output
func outputFormsMap() map[string](func(Configuration) string) {
	return map[string](func(Configuration) string){
		"ansible": AnsibleOutputFor,
		"iptables": IpTablesOutputFor,
	}
}

// Returns the function that generates the output from the output type
func ChooseOutputForm(form string) func(Configuration) string {
	if f, found := outputFormsMap()[form]; !found {
		panic(fmt.Sprintf("Key '%s' not found", form))
	} else {
		return f
	}
}

// Parse the command line for options and argumets the return the
// spec file and the output format type
func getWorkToDoFromCmdline() (*string, *string) {
	help := flag.Bool("help", false, "Show help")
	inSpec := flag.String("input-spec", "", "Input Load Balancer spec (JSON)")
	outputType := flag.String("output-type", "iptables",
	"One of: ansible|iptables|iptables-rules (pending)|iptables-run (pending)..")

	flag.Parse()

	if *help || (len(*inSpec) == 0) {
		flag.Usage()
		os.Exit(0)
	}
	return inSpec, outputType
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
	inSpec, outputType := getWorkToDoFromCmdline()	

	cfg := getLBSpecFromJSONFile(*inSpec)
	//fmt.Println(cfg.Pomaloba)
	//fmt.Println(cfg.App)
	//fmt.Println(cfg.Endpoints)
	//fmt.Println(cfg.Destinations)

	f := ChooseOutputForm(*outputType)
	fmt.Print(f(cfg))
}

/* Local Variables: */
/* mode: Go */
/* tab-width: 3 */
/* End: */
