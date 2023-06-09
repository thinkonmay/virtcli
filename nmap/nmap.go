package nmap

import (
	"fmt"
	"os/exec"
	"strings"
	"test/model"
)

func FindIPMac(subnet string) (result map[string]string) {
	result = map[string]string{}
	out,err := exec.Command("nmap","-sn",subnet).Output()
	if err != nil {
		return 
	}

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
			fmt.Printf("found %s\n", string(ipline))
			words := strings.Split(string(ipline), " ")
			for i2, v2 := range words {
				if v2 == "for" && words[i2-1] == "report" {
					ip = words[i2+1]
				}
			}
		}

		if ip != "" {
			result[mac]=ip
		}
	}

	return
}



func FindDomainIPs(dom model.Domain) []string {

	macs := []string{}
	for _, i2 := range dom.Interfaces {
		macs = append(macs, *i2.Mac.Address)
	}

	ips := []string{}
	for k, v := range FindIPMac("192.168.1.*") {
		for _, v2 := range macs {
			if v2 == k {
				ips = append(ips, v)
			}
		}
	}

	return ips
}