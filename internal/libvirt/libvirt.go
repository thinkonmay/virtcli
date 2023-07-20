package libvirt

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"test/internal/network/ovs"
	"test/model"
	"test/utils/arp"
	"time"

	"github.com/digitalocean/go-libvirt"
	"gopkg.in/yaml.v3"
)

type Libvirt struct {
	Version string 
	conn *libvirt.Libvirt

	vswitch *ovs.OpenVSwitch
}

var (
	ifwhitelist = []string{"enp0s25","enp11s0","enp5s0"}
)

func NewLibvirt() *Libvirt {
	ret := &Libvirt{
		vswitch: ovs.NewOVS(),
	}

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


func (lv *Libvirt)deleteDisks(path string) error {
	if strings.Contains(path, "do-not-delete") {
		return fmt.Errorf("resource name contain do-not-delete tag")
	}

	dev,_,_ := lv.conn.ConnectListAllStoragePools(1,libvirt.ConnectListStoragePoolsActive)

	for _, nd := range dev {
		err := lv.conn.StoragePoolRefresh(nd,0)
		if err != nil {
			continue
		}

		vols,_,err := lv.conn.StoragePoolListAllVolumes(nd,1,0)
		if err != nil {
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
			if vl.Path == path && vl.Format.Type == "qcow2"{
				return lv.conn.StorageVolDelete(sv,libvirt.StorageVolDeleteNormal)
			}
		}
	}

	return nil
}
func (lv *Libvirt)ListDisks() []model.Volume{
	dev,_,_ := lv.conn.ConnectListAllStoragePools(1,libvirt.ConnectListStoragePoolsActive)

	
	ret := []model.Volume{}
	for _, nd := range dev {
		if strings.Contains(strings.ToLower(nd.Name),"ignore") {
			continue
		}

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


func (lv *Libvirt)ListDomainIPs(dom model.Domain) []string { // TODO

	flags := libvirt.ConnectListDomainsActive 
	domains, _, err := lv.conn.ConnectListAllDomains(1, flags)
	if err != nil {
		return []string{}
	}

	virtdom := libvirt.Domain{Name: "unknown"}
	for _, d := range domains {
		if *dom.Name == d.Name {
			virtdom = d
		}
	}

	if virtdom.Name == "unknown" {
		return []string{}
	}

	return arp.FindDomainIPs(dom)
}


func backingChain(vols []model.Volume, target model.Volume) *model.BackingStore {
	var backing *model.BackingStore = nil 

	for _,v := range vols {
		if v.Path != target.Path || v.Backing == nil {
			continue
		}

		backingChild := model.Volume{}
		for _, v2 := range vols {
			if v.Backing.Path == v2.Path {
				backingChild = v2
			}
		}


		backing = &model.BackingStore{
			Type: "file",
			Format: &struct{Type string "xml:\"type,attr\""}{
				Type: "qcow2",
			},
			Source: &struct{File string "xml:\"file,attr\""}{
				File: v.Backing.Path,
			},
			BackingStore: backingChain(vols,backingChild),
		}
	}

	return backing
}
func VolType(vols []model.Volume, target model.Volume) string {
	if  strings.Contains(strings.ToLower(target.Path),"os"){
		return "os"
	} else if strings.Contains(strings.ToLower(target.Path),"game") || 
			  strings.Contains(strings.ToLower(target.Path),"app") {
		return "app"
	}

	for _, v := range vols {
		if v.Path != target.Path || v.Backing == nil {
			continue
		}

		backingChild := model.Volume{}
		for _, v2 := range vols {
			if v.Backing.Path == v2.Path {
				backingChild = v2
			}
		}


		return VolType(vols, backingChild)
	}

	return "unknown"
}


func (lv *Libvirt)CreateVM(vcpus int,
							ram int,
							gpus []model.GPU,
							vols []model.Volume,
							) (string,error) {
	if vcpus % 2 == 1 {
		return "",fmt.Errorf("vcpus should not be odd")
	}

	file,err := os.OpenFile("./model/data/vm.yaml",os.O_RDWR,0755)
	if err != nil {
		panic(err)
	}

	data,err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	dom := model.Domain{}
	yaml.Unmarshal(data, &dom)

	file.Close()
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

	voldb := lv.ListDisks()
	dom.Disk = []model.Disk{}
	for i,d := range vols {
		dev := "hda"
		if i == 1 {
			dev = "hdb"
		}

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
				Dev: dev,
				Bus: "ide",
			},
			Address: nil,
			Type: d.Type,
			Device: "disk",
			BackingStore: backingChain(voldb,d),
		})
	}

	dom.Interfaces = []model.Interface{}
	iface,err := lv.vswitch.CreateInterface()
	if err != nil {
		return "", err
	}

	dom.Interfaces = append(dom.Interfaces, *iface)

	dom.Memory.Value        = ram * 1024 * 1024
	dom.CurrentMemory.Value = ram * 1024 * 1024

	dom.VCpu.Value 			= vcpus
	dom.Cpu.Topology.Socket = 1
	dom.Cpu.Topology.Cores  = vcpus / 2
	dom.Cpu.Topology.Thread = 2

	xml := dom.ToString()
	result,err := lv.conn.DomainDefineXMLFlags(xml,libvirt.DomainDefineValidate)
	if err != nil {
		return "",err
	}


	err = lv.conn.DomainCreate(result)
	if err != nil {
		return "",err
	}
	return string(result.Name),nil
}

func (lv *Libvirt)StopVM(name string) (error) {
	flags := libvirt.ConnectListDomainsActive | libvirt.ConnectListDomainsInactive
	doms,_,err := lv.conn.ConnectListAllDomains(1,flags)
	if err != nil {
		return err
	}

	dom := libvirt.Domain{Name: "null"}
	for _, d := range doms {
		if d.Name == name {
			dom = d
		}
	}

	if dom.Name == "null" {
		return fmt.Errorf("unknown VM name")
	}

	start := time.Now()
	for {
		err = lv.conn.DomainShutdown(dom)
		if err == nil || time.Now().UnixMilli() - start.UnixMilli() > 10 * 1000 {
			break
		}

		time.Sleep(1 * time.Second)
	}

	return err
}
func (lv *Libvirt)StartVM(name string,
						  gpus []model.GPU) (error) {
	flags := libvirt.ConnectListDomainsActive | libvirt.ConnectListDomainsInactive
	doms,_,err := lv.conn.ConnectListAllDomains(1,flags)
	if err != nil {
		return err
	}

	model_domain := model.Domain{}
	old_dom := libvirt.Domain{Name: "null"}

	models := lv.ListDomains()
	for _, d := range doms {
		if d.Name == name {
			old_dom = d
			for _, d2 := range models {
				if d.Name == *d2.Name {
					model_domain = d2
				}
			}
		}
	}

	if old_dom.Name == "null" {
		return fmt.Errorf("unknown VM name")
	}

	Attach := []model.HostDev{}
	for _, nd := range gpus {
		for _, v := range nd.Capability.IommuGroup.Address {
			Attach = append(Attach, model.HostDev{
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

	model_domain.Interfaces = []model.Interface{}
	lv.conn.DomainUndefine(old_dom)

	model_domain.Uuid = nil
	model_domain.Hostdevs = Attach

	model_domain.Interfaces = []model.Interface{}
	iface,err := lv.vswitch.CreateInterface()
	if err != nil {
		return err
	}
	model_domain.Interfaces = append(model_domain.Interfaces, *iface)


	new_dom,err := lv.conn.DomainDefineXML(model_domain.ToString())
	if err != nil {
		return err
	}

	start := time.Now()
	for {
		err = lv.conn.DomainCreate(new_dom)
		if err == nil || time.Now().UnixMilli() - start.UnixMilli() > 10 * 1000 {
			break
		}

		time.Sleep(1 * time.Second)
	}

	return err
}

func (lv *Libvirt)DeleteVM(name string,running bool) (error) {
	if strings.Contains(name, "do-not-delete") {
		return fmt.Errorf("resource name contain do-not-delete tag")
	}

	flags := libvirt.ConnectListDomainsActive | libvirt.ConnectListDomainsInactive
	doms,_,err := lv.conn.ConnectListAllDomains(1,flags)
	if err != nil {
		return err
	}

	dom := libvirt.Domain{Name: "null"}
	for _, d := range doms {
		if d.Name == name {
			dom = d
		}
	}

	if dom.Name == "null" {
		return fmt.Errorf("unknown VM name")
	}

	if running {
		err = lv.StopVM(name)
		if err != nil {
			return err
		}
	}

	desc,err := lv.conn.DomainGetXMLDesc(dom,libvirt.DomainXMLSecure)
	if err != nil {
		return err
	}

	dommodel := model.Domain{}
	err = dommodel.Parse([]byte(desc))
	if err != nil {
		return err
	}

	for _, d := range dommodel.Disk {
		if d.Source == nil || d.Driver.Type != "qcow2" {
			continue
		} else if err := lv.deleteDisks(d.Source.File); err != nil {
			return err
		}
	}

	start := time.Now().UnixMilli()
	for {
		lv.conn.DomainUndefine(dom)
		doms,_,err = lv.conn.ConnectListAllDomains(1,flags)
		if err != nil {
			return err
		}

		found := false
		for _, d := range doms {
			if d.Name == name {
				found = true
			}
		}

		if !found {
			return nil
		} else if time.Now().UnixMilli() - start > 10 * time.Second.Milliseconds() {
			return fmt.Errorf("timeout")
		}

		time.Sleep(500 * time.Millisecond)
	}
}


// // stats,err := l.ConnectGetAllDomainStats(domains,0,libvirt.ConnectGetAllDomainsStatsActive)
// // fmt.Printf("%v\n",stats)
