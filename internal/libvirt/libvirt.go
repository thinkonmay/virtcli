package libvirt

import (
	"encoding/xml"
	"fmt"
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




func (lv *Libvirt)CreateVM(	id string,
							vcpus int,
							ram int,
							gpus []model.GPU,
							vols []model.Disk,
							) (string,error) {
	if vcpus % 2 == 1 {
		return "",fmt.Errorf("vcpus should not be odd")
	}

	doms,_,err := lv.conn.ConnectListAllDomains(0,libvirt.ConnectListDomainsActive|libvirt.ConnectListDomainsInactive)
	if err != nil {
		return "",err
	}

	for _, d := range doms {
		if d.Name == id {
			lv.conn.DomainDestroy(d)
			lv.conn.DomainUndefine(d)
		}
	}

	dom := model.Domain{}
	yaml.Unmarshal([]byte(libvirtVM), &dom)

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


	iface,err := lv.vswitch.CreateInterface()
	if err != nil {
		return "", err
	}

	dom.Name 				= &id
	dom.Uuid 				= nil
	dom.Disk 				= vols
	dom.Interfaces 			= []model.Interface{*iface}

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
		if err != nil {
			if strings.Contains(err.Error(),"domain is not running") {
				return nil
			}

			fmt.Fprintf(os.Stderr, "failed to shutdown %s\n",err.Error())
			time.Sleep(1 * time.Second)
		}

		if time.Now().UnixMilli() - start.UnixMilli() > 30 * 1000 {
			return fmt.Errorf("timeout shutting down VM %s",name)
		} 
	}
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
		return nil
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
			return fmt.Errorf("timeout delete VM %s",name)
		}

		time.Sleep(500 * time.Millisecond)
	}
}



func (lv *Libvirt) CreateTempPool(path string) (*libvirt.StoragePool,error) {
	now := fmt.Sprintf("%d",time.Now().UnixMilli())
	result,_ := xml.Marshal(model.StoragePool{
		Type: "dir",
		Name: now,
		Path: path,
	})
	pool,err := lv.conn.StoragePoolDefineXML(string(result),0)
	if err != nil {
		return nil, err
	}
	return &pool,nil
}


func (lv *Libvirt) RemovePool(pool libvirt.StoragePool) (error) {
	return lv.conn.StoragePoolUndefine(pool)
}