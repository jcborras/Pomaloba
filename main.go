
package main

import (
	"flag"
	"fmt"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
)

type EndPoint struct {
	IP string
	Port uint16
}

type Configuration struct {
	Pomaloba int `json:"pomaloba"`
	App string `json:"app"`
	Endpoints []EndPoint `json:"endpoints"`
	Destinations []EndPoint `json:"destinations"`
}

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func main() {

	help := flag.Bool("help", false, "Show help")
	inSpec := flag.String("input-spec", "", "Input LB specification")
	// outputType := flag.String("output-spec", "iptables", "Output style: iptables|..")
	flag.Parse()
	
	if (*help || (len(*inSpec) == 0)){
		flag.Usage()
		os.Exit(0)
	}
	dat, err1 := os.ReadFile(*inSpec)
	check(err1)
	var cfg Configuration
	err2 := json.Unmarshal(dat, &cfg)
	check(err2)
	//fmt.Println(cfg.Pomaloba)
	//fmt.Println(cfg.App)
	//fmt.Println(cfg.Endpoints)
	//fmt.Println(cfg.Destinations)

	outStr := " --table nat --append PREROUTING" + 
                  " --protocol tcp --destination %s --dport %d" + 
		  " --match statistic --mode random --probability %.5f" + 
		  " --jump DNAT --to-destination %s:%d"

	for _, ep := range cfg.Endpoints {
		for j, dst := range cfg.Destinations {
			p := 1.0/float32(len(cfg.Destinations)-j)
			cmdStr := fmt.Sprintf(outStr, ep.IP, ep.Port, p, dst.IP, dst.Port)
			cmd := strings.Split(cmdStr, " ")
			exec.Command("iptables", cmd...)
			// TODO: check output
		}
	}
}
