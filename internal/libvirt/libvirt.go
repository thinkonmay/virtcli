package libvirt

import (
	"fmt"
	"log"
	"net"
	"strings"
	"test/model"
	"time"

	"github.com/digitalocean/go-libvirt"
)

type Libvirt struct {
	Version string 
	conn *libvirt.Libvirt

}

func NewLibvirt() *Libvirt {
	ret := &Libvirt{}

	c, err := net.DialTimeout("unix", "/var/run/libvirt/libvirt-sock", 2*time.Second)
	if err != nil {
		log.Fatalf("failed to dial libvirt: %v", err)
	}

	ret.conn = libvirt.New(c)
	if err := ret.conn.Connect(); err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	ret.Version, err = ret.conn.Version()
	if err != nil {
		log.Fatalf("failed to retrieve libvirt version: %v", err)
	}


	return ret
}




func (lv *Libvirt)ListDomains() []model.Domain{
	flags := libvirt.ConnectListDomainsActive | libvirt.ConnectListDomainsInactive
	domains, _, err := lv.conn.ConnectListAllDomains(1, flags)
	if err != nil {
		log.Fatalf("failed to retrieve domains: %v", err)
	}


	ret := []model.Domain{}
	for _, d := range domains {
		desc,err := lv.conn.DomainGetXMLDesc(d,libvirt.DomainXMLSecure)
		if err != nil {
			continue
		}

		dom := model.Domain{}
		dom.Parse([]byte(desc))
		ret = append(ret, dom)
	}


	return ret

}

func (lv *Libvirt)ListGPUs() []model.GPU{
	dev,_,_ := lv.conn.ConnectListAllNodeDevices(1,0)

	ret := []model.GPU{}
	for _, nd := range dev {
		desc,err := lv.conn.NodeDeviceGetXMLDesc(nd.Name,0)
		if err != nil {
			continue
		}

		gpu := model.GPU{}
		gpu.Parse([]byte(desc))

		vendor := strings.ToLower(gpu.Capability.Vendor.Val)
		if !strings.Contains(vendor,"nvidia") {
			continue
		} 
		product := strings.ToLower(gpu.Capability.Product.Val)
		if strings.Contains(product,"audio") {
			continue
		}

		ret = append(ret, gpu)
	}

	return ret
}

func (lv *Libvirt)CreateVM(dom model.Domain,gpus []model.GPU) (string,error) {
	name := fmt.Sprintf("%d", time.Now().Nanosecond())
	dom.Name = &name
	dom.Uuid = nil


	dom.Hostdevs = []model.HostDev{}
	for _, nd := range gpus {
		for _, v := range nd.Capability.IommuGroup.Address {
			dom.Hostdevs = append(dom.Hostdevs, model.HostDev{
				Mode: "subsystem",
				Type: "pci",
				Managed: "yes",
				SourceAddress: &struct{
					Domain string "xml:\"domain,attr\""; 
					Bus string "xml:\"bus,attr\""; 
					Slot string "xml:\"slot,attr\""; 
					Function string "xml:\"function,attr\"";
				}{
					Domain : v.Domain,
					Bus : v.Bus,
					Slot : v.Slot,
					Function : v.Function,
				},
			})
		}
	}

	xml := dom.ToString()
	result,err := lv.conn.DomainCreateXML(xml,libvirt.DomainNone)
	if err != nil {
		return "",err
	}

	return string(result.Name),nil
}

// // stats,err := l.ConnectGetAllDomainStats(domains,0,libvirt.ConnectGetAllDomainsStatsActive)
// // fmt.Printf("%v\n",stats)

// // dev,_,err := l.ConnectListAllNodeDevices(1,1)
// // fmt.Printf("%v\n",dev)

// networks, _, err := l.ConnectListAllNetworks(1, libvirt.ConnectListNetworksActive)
// for _, n := range networks {
// 	desc, _ := l.NetworkGetXMLDesc(n, 0)
// 	fmt.Printf("%v\n", desc)
// }

// ifaces, _, err := l.ConnectListAllInterfaces(1, libvirt.ConnectListInterfacesActive)
// for _, i2 := range ifaces {
// 	desc, _ := l.InterfaceGetXMLDesc(i2, 0)
// 	fmt.Printf("%v\n", desc)
// }

// pool, _, err := l.ConnectListAllStoragePools(1, libvirt.ConnectListStoragePoolsActive)
// for _, sp := range pool {
// 	l.StoragePoolRefresh(sp, 0)
// 	vol, _, err := l.StoragePoolListAllVolumes(sp, 1, 0)
// 	if err != nil {
// 	}

// 	for _, sv := range vol {
// 		desc, _ := l.StorageVolGetXMLDesc(sv, 0)
// 		fmt.Printf("%v\n", desc)
// 	}
// 	fmt.Printf("%v\n", vol)
// }

// state, err := l.DomainState("vm5")
// fmt.Printf("%s\n", state)

// if err := l.Disconnect(); err != nil {
// 	log.Fatalf("failed to disconnect: %v", err)
// }