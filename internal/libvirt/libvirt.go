package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/digitalocean/go-libvirt"
)

func main() {
	// This dials libvirt on the local machine, but you can substitute the first
	// two parameters with "tcp", "<ip address>:<port>" to connect to libvirt on
	// a remote machine.
	c, err := net.DialTimeout("unix", "/var/run/libvirt/libvirt-sock", 2*time.Second)
	if err != nil {
		log.Fatalf("failed to dial libvirt: %v", err)
	}

	l := libvirt.New(c)
	if err := l.Connect(); err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	v, err := l.Version()
	if err != nil {
		log.Fatalf("failed to retrieve libvirt version: %v", err)
	}
	fmt.Println("Version:", v)

	flags := libvirt.ConnectListDomainsActive | libvirt.ConnectListDomainsInactive
	domains, _, err := l.ConnectListAllDomains(1, flags)
	if err != nil {
		log.Fatalf("failed to retrieve domains: %v", err)
	}

	fmt.Println("ID\tName\t\tUUID")
	fmt.Printf("--------------------------------------------------------\n")
	for _, d := range domains {
		// _,_ = l.DomainGetXMLDesc(d,libvirt.DomainXMLSecure)
		// l.DomainCreateXML()
		// name,_ := l.DomainGetHostname(d,libvirt.DomainGetHostnameLease)
		// param,_ := l.DomainGetGuestInfo(d,1,1)
		fmt.Printf("%d\t%s\t%x\n", d.ID, d.Name, d.UUID)
	}

	// stats,err := l.ConnectGetAllDomainStats(domains,0,libvirt.ConnectGetAllDomainsStatsActive)
	// fmt.Printf("%v\n",stats)

	// dev,_,err := l.ConnectListAllNodeDevices(1,1)
	// fmt.Printf("%v\n",dev)

	networks, _, err := l.ConnectListAllNetworks(1, libvirt.ConnectListNetworksActive)
	for _, n := range networks {
		desc, _ := l.NetworkGetXMLDesc(n, 0)
		fmt.Printf("%v\n", desc)
	}

	ifaces, _, err := l.ConnectListAllInterfaces(1, libvirt.ConnectListInterfacesActive)
	for _, i2 := range ifaces {
		desc, _ := l.InterfaceGetXMLDesc(i2, 0)
		fmt.Printf("%v\n", desc)
	}

	pool, _, err := l.ConnectListAllStoragePools(1, libvirt.ConnectListStoragePoolsActive)
	for _, sp := range pool {
		l.StoragePoolRefresh(sp, 0)
		vol, _, err := l.StoragePoolListAllVolumes(sp, 1, 0)
		if err != nil {
		}

		for _, sv := range vol {
			desc, _ := l.StorageVolGetXMLDesc(sv, 0)
			fmt.Printf("%v\n", desc)
		}
		fmt.Printf("%v\n", vol)
	}

	state, err := l.DomainState("vm5")
	fmt.Printf("%s\n", state)

	if err := l.Disconnect(); err != nil {
		log.Fatalf("failed to disconnect: %v", err)
	}
}
