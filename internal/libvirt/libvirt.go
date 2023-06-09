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

func (lv *Libvirt)ListIfaces() []model.Iface{
	dev,_,_ := lv.conn.ConnectListAllInterfaces(1,libvirt.ConnectListInterfacesActive)

	
	ret := []model.Iface{}
	for _, nd := range dev {
		xml,err := lv.conn.InterfaceGetXMLDesc(nd,0)
		if err != nil {
			continue
		}
		iface := &model.Iface{}
		iface.Parse(xml)
		ret = append(ret, *iface)
	}

	return ret
}

func (lv *Libvirt)ListDisks() []model.Volume{
	dev,_,_ := lv.conn.ConnectListAllStoragePools(1,libvirt.ConnectListStoragePoolsActive)

	
	ret := []model.Volume{}
	for _, nd := range dev {
		err:= lv.conn.StoragePoolRefresh(nd,0)
		if err != nil {
			fmt.Printf("%s\n",err.Error())
			continue
		}
		vols,_,err := lv.conn.StoragePoolListAllVolumes(nd,1,0)
		if err != nil {
			fmt.Printf("%s\n",err.Error())
			continue
		}

		for _, sv := range vols {
			xml,err := lv.conn.StorageVolGetXMLDesc(sv,0)
			if err != nil {
				fmt.Printf("%s\n",err.Error())
				continue
			}

			vl := model.Volume{}
			vl.Parse(xml)
			if vl.Format.Type != "qcow2" {
				continue
			}

			ret = append(ret, vl)
		}
	}

	return ret
}

func (lv *Libvirt)CreateVM(dom model.Domain,
							gpus []model.GPU,
							vols []model.Volume,
							ifs  []model.Iface,
							) (string,error) {
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

	dom.Disk = []model.Disk{}
	for _,d := range vols {
		disk := "disk"
		dom.Disk = append(dom.Disk, model.Disk{
			Driver: &struct{Name string "xml:\"name,attr\""; Type string "xml:\"type,attr\""}{
				Name: "qemu",
				Type: d.Format.Type,
			},
			Source: &struct{File string "xml:\"file,attr\""; Index int "xml:\"index,attr\""}{
				File: d.Path,
				Index: 1,
			},
			Target: &struct{Dev string "xml:\"dev,attr\""; Bus string "xml:\"bus,attr\""}{
				Dev: "hda",
				Bus: "ide",
			},
			Address: nil,
			Type: &d.Type,
			Device: &disk,
		})
	}

	dom.Interfaces = []model.Interface{}
	for _,d := range ifs {
		e1000 := "e1000"
		dom.Interfaces = append(dom.Interfaces, model.Interface{
			Type: "direct",
			Source: &struct{Dev string "xml:\"dev,attr\""; Mode string "xml:\"mode,attr\""}{
				Dev: *d.Name,
				Mode: "bridge",
			},
			Model: &struct{Type *string "xml:\"type,attr\""}{
				Type: &e1000,
			},
			// Mac: &struct{Address *string "xml:\"address,attr\""}{
			// 	Address: d.Mac.Address,
			// },
		})
	}

	xml := dom.ToString()
	result,err := lv.conn.DomainCreateXML(xml,libvirt.DomainNone)
	if err != nil {
		return "",err
	}

	return string(result.Name),nil
}


func (lv *Libvirt)DeleteVM(name string) (error) {
	flags := libvirt.ConnectListDomainsActive | libvirt.ConnectListDomainsInactive
	doms,_,err := lv.conn.ConnectListAllDomains(1,flags)
	if err != nil {
		return err
	}

	var dom *libvirt.Domain = nil
	for _, d := range doms {
		if d.Name == name {
			dom = &d
		}
	}

	if dom == nil {
		return fmt.Errorf("unknown VM name")
	}


	return lv.conn.DomainShutdown(*dom)
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