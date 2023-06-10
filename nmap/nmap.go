package nmap

import (
	"os/exec"
	"strings"
	"sync"
	"test/model"
	"time"
)


var database = map[string]string{}
var mut = &sync.Mutex{}

func init() {
	findIPMac("192.168.1.*")
	go func ()  {
		for {
			time.Sleep(5 * time.Second)
			findIPMac("192.168.1.*")
		}
	}()
}
func FindIPMac() map[string]string {
	return database
}
func findIPMac(subnet string) () {
	out,err := exec.Command("nmap","-sn",subnet).Output()
	if err != nil {
		return 
	}


	mut.Lock()
	defer mut.Unlock()
	database = map[string]string{}
	lines := strings.Split(string(out), "\n")
	for i, v := range lines {
		if v == "" {
			continue
		}

		mac,ip := "",""
		words := strings.Split(string(v), " ")
		for i2, v2 := range words {
			if v2 == "Address:" && words[i2-1] == "MAC" {
				mac = words[i2+1]
			}
		}

		if mac != "" {
			ipline := lines[i-2]
			words := strings.Split(string(ipline), " ")
			for i2, v2 := range words {
				if v2 == "for" && words[i2-1] == "report" {
					ip = words[i2+1]
				}
			}
		}

		if ip != "" {
			database[mac]=ip
		}
	}

	return
}






func FindDomainIPs(dom model.Domain) []string {

	ips := []string{}
	macs := []string{}
	for _, i2 := range dom.Interfaces {
		macs = append(macs, *i2.Mac.Address)
	}

	mut.Lock()
	defer mut.Unlock()
	for k, v := range database {
		for _, v2 := range macs {
			if v2 == k {
				ips = append(ips, v)
			}
		}
	}

	return ips
}