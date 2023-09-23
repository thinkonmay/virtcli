package libvirt

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	qemuhypervisor "test/internal/libvirt/qemu"
	"test/internal/network"
	libvirtnetwork "test/internal/network/libvirt"
	"test/internal/network/ovs"
	"test/model"
	"time"

	"github.com/digitalocean/go-libvirt"
	"github.com/digitalocean/go-libvirt/socket/dialers"
	"github.com/digitalocean/go-qemu/qemu"
	"gopkg.in/yaml.v3"
)

type Libvirt struct {
	Version string
	conn    *libvirt.Libvirt

	network network.Network
	qemu    qemuhypervisor.QEMUHypervisor
}

func NewLibvirt() *Libvirt {
	ret := &Libvirt{
		qemu:    *qemuhypervisor.NewQEMUHypervisor(),
	}

	ovsif := os.Getenv("OVS_IFACE")
	if ovsif == "" {
		ret.network = libvirtnetwork.NewLibvirtNetwork()
	} else {
		ret.network = ovs.NewOVS(ovsif)
	}

	c, err := net.DialTimeout("unix", "/var/run/libvirt/libvirt-sock", 2*time.Second)
	if err != nil {
		log.Fatalf("failed to dial libvirt: %v", err)
	}

	ret.conn = libvirt.NewWithDialer(dialers.NewAlreadyConnected(c))
	if err := ret.conn.Connect(); err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	return ret
}

func (lv *Libvirt) ListDomains() []model.Domain {
	flags := libvirt.ConnectListDomainsActive | libvirt.ConnectListDomainsInactive
	domains, _, err := lv.conn.ConnectListAllDomains(1, flags)
	if err != nil {
		log.Fatalf("failed to retrieve domains: %v", err)
	}

	ret := []model.Domain{}
	statuses := lv.qemu.ListDomainWithStatus()
	for _, d := range domains {
		desc, err := lv.conn.DomainGetXMLDesc(d, libvirt.DomainXMLSecure)
		if err != nil {
			continue
		}

		dom := model.Domain{}
		dom.Parse([]byte(desc))
		for _, d2 := range statuses {
			if d2.Name == d.Name {
				status := d2.Status.String()
				dom.Status = &status
			}
		}

		ret = append(ret, dom)
	}

	return ret

}

func (lv *Libvirt) ListGPUs() []model.GPU {
	dev, _, _ := lv.conn.ConnectListAllNodeDevices(1, 0)

	ret := []model.GPU{}
	for _, nd := range dev {
		desc, err := lv.conn.NodeDeviceGetXMLDesc(nd.Name, 0)
		if err != nil {
			continue
		}

		gpu := model.GPU{}
		gpu.Parse([]byte(desc))

		vendor := strings.ToLower(gpu.Capability.Vendor.Val)
		if !strings.Contains(vendor, "nvidia") {
			continue
		}
		product := strings.ToLower(gpu.Capability.Product.Val)
		if strings.Contains(product, "audio") {
			continue
		}

		ret = append(ret, gpu)
	}

	return ret
}

func (lv *Libvirt) ListDomainIPs(dom model.Domain) []string { // TODO
	return lv.network.FindDomainIPs(dom)
}
















func (lv *Libvirt) CreateVM(id string,
	vcpus int,
	ram int,
	gpus []model.GPU,
	vols []model.Disk,
) (string, error) {
	dom := model.Domain{}
	yaml.Unmarshal([]byte(libvirtVM), &dom)
	dom.Hostdevs = []model.HostDev{}
	for _, nd := range gpus {
		for _, v := range nd.Capability.IommuGroup.Address {
			dom.Hostdevs = append(dom.Hostdevs, model.HostDev{
				Mode:    "subsystem",
				Type:    "pci",
				Managed: "yes",
				SourceAddress: &struct {
					Domain   string "xml:\"domain,attr\""
					Bus      string "xml:\"bus,attr\""
					Slot     string "xml:\"slot,attr\""
					Function string "xml:\"function,attr\""
				}{
					Domain:   v.Domain,
					Bus:      v.Bus,
					Slot:     v.Slot,
					Function: v.Function,
				},
			})
		}
	}

	iface, err := lv.network.CreateInterface()
	if err != nil {
		return "", err
	}

	dom.Name = &id
	dom.Uuid = nil
	dom.Disk = vols
	dom.Interfaces = []model.Interface{*iface}

	dom.Memory.Value = ram * 1024 * 1024
	dom.CurrentMemory.Value = ram * 1024 * 1024

	dom.VCpu.Value = vcpus
	dom.Cpu.Topology.Socket = 1
	dom.Cpu.Topology.Thread = 1
	dom.Cpu.Topology.Cores = vcpus 

	xml := dom.ToString()
	result, err := lv.conn.DomainCreateXML(xml, libvirt.DomainStartValidate)
	if err != nil {
		return "", err
	}




	time.Sleep(10 * time.Second)
	statuses := lv.qemu.ListDomainWithStatus()
	for _, d := range statuses {
		if d.Name == id && d.Status != qemu.StatusRunning{
			lv.conn.DomainDestroy(result)
			lv.conn.DomainUndefine(result)
			return "",fmt.Errorf("domain %s failed to start after 30s",id)
		}
	}
	return string(result.Name), nil
}

func (lv *Libvirt) DeleteVM(name string) error {
	if strings.Contains(name, "do-not-delete") {
		return nil
	}

	flags := libvirt.ConnectListDomainsActive 
	doms, _, err := lv.conn.ConnectListAllDomains(1, flags)
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



	
	lv.conn.DomainShutdown(dom)
	time.Sleep(10 * time.Second)
	statuses := lv.qemu.ListDomainWithStatus()
	for _, d := range statuses {
		if d.Name == dom.Name {
			lv.conn.DomainDestroy(dom)
			lv.conn.DomainUndefine(dom)
		}
	}

	return nil
}



func (lv *Libvirt) GetCPUPinning(count int) *model.CPU {
	ret,available := []cpu.HostCore{},[]cpu.HostCore{}

	doms := lv.ListDomains()
	Topo,err := cpu.GetCPUPinning()
	for _,cpu := range Topo.CPUs {
		add := true
		for _,dom := range doms {
			for _,pin := dom.Cpu.Vcpupin {
				if pin == cpu.CPU {
					add = false
				}
			}
		}

		if !add {
			continue;
		}

		available = append(available,cpu)
	}


	sockets := map[int]struct{
		core map[int]int
	}{}

	for _,core := range available {


	}



	return &mode.CPU{

	}
}