package libvirt

import (
	"fmt"
	"log"
	"net"
	"os"
	// "os/exec"
	"sort"
	"strconv"
	"strings"
	"test/host/cpu"
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
		err = dom.Parse([]byte(desc))
		if err != nil {
			panic(err)
		}
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
		err = gpu.Parse([]byte(desc))
		if err != nil {
			panic(err)
		}

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
	VDriver bool,
	HideVM  bool,
) (string, error) {
	dom := model.Domain{}
	err := yaml.Unmarshal([]byte(libvirtVM), &dom)
	if err != nil { return "", err }

	dom.Name = &id
	dom.Disk = vols
	dom.Memory.Value = ram * 1024 * 1024
	dom.CurrentMemory.Value = ram * 1024 * 1024
	dom.VCpu.Value = vcpus

	dom.Cpu.Topology.Socket = 1
	dom.Cpu.Topology.Thread = 2
	dom.Cpu.Topology.Cores  = vcpus / 2

	dom.Hostdevs = []model.HostDev{}
	dom.Vcpupin  = []model.Vcpupin{}

	for _, nd := range gpus {
		dom.Vcpupin,err = lv.GetCPUPinning(vcpus,*nd.Capability.Numa.Node)
		if err != nil {
			return "", err
		}
		dom.NumaTune = &model.NumaTune{
			Memory: struct {
				Mode string `xml:"mode,attr"`
				Nodeset int `xml:"nodeset,attr"`
			}{
				Mode: "strict",
				Nodeset: *nd.Capability.Numa.Node,
			},
		}

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
		})}
	}

	driver := "e1000e"
	if VDriver {
		driver = "virtio"
	}

	iface, err := lv.network.CreateInterface(driver)
	dom.Interfaces = []model.Interface{*iface}
	if err != nil {
		return "", err
	}

	if !HideVM {
		dom.Cpu.Feature = nil
		dom.Features.Kvm = nil
		dom.OS.Smbios = nil
	}

	xml := dom.ToString()
	fmt.Println(xml)
	result, err := lv.conn.DomainCreateXML(xml, libvirt.DomainStartValidate)
	if err != nil {
		return "", fmt.Errorf("error starting VM: %s",err.Error())
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

	// for _,v := range Vcpupin {
	// 	fmt.Printf("pin host cpu %d to guest cpu %d",v.Cpuset,v.Vcpu)
	// 	result,err := exec.Command("virsh","vcpupin",
	// 		"--vcpu",fmt.Sprintf("%d",v.Vcpu),
	// 		"--cpulist",fmt.Sprintf("%d",v.Cpuset),
	// 		"--live").Output()
	// 	if err != nil {
	// 		fmt.Println(err.Error())
	// 	} else {
	// 		fmt.Println(string(result))
	// 	}
	// }
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



func (lv *Libvirt) GetCPUPinning(count int,
								 numa_node int,
								 ) ([]model.Vcpupin,error) {
	available := []cpu.HostCore{}

	doms := lv.ListDomains()
	Topo,err := cpu.GetHostTopology()
	if err != nil {
		return nil,err
	}

	for _,cpu := range Topo.CPUs {
		add := true
		for _,dom := range doms {
			if *dom.Status != "StatusRunning" {
				continue
			}

			for _,pin := range dom.Vcpupin {
				if fmt.Sprintf("%d",pin.Cpuset) == cpu.CPU {
					add = false
				}
			}
		}

		if !add {
			continue;
		}

		available = append(available,cpu)
	}


	all := map[string]map[string][]string{}
	max := map[string]map[string][]string{}

	for _,core := range available {
		if all[core.Socket] == nil { all[core.Socket] = map[string][]string{ core.Core : {} } }
		all[core.Socket][core.Core] = append(all[core.Socket][core.Core], core.CPU)
	}
	for _,core := range Topo.CPUs {
		if max[core.Socket] == nil { max[core.Socket] = map[string][]string{ core.Core : {} } }
		max[core.Socket][core.Core] = append(max[core.Socket][core.Core], core.CPU)
	}

	sockets := make([]string, 0, len(max))
	for k := range max { sockets = append(sockets, k) }
	
	cores := make([]string, 0, len(max[sockets[0]]))
	for k := range max[sockets[0]] { cores = append(cores, k) }
	sort.Slice(cores, func(i, j int) bool {
		a,_ := strconv.ParseInt(cores[i],10,32)
		b,_ := strconv.ParseInt(cores[j],10,32)
		return a < b
	})

	thread_per_core := len(max[sockets[0]][cores[0]])
	if count % thread_per_core != 0 {
		return nil,fmt.Errorf("cpu count not even")
	}



	vcpupin := []model.Vcpupin{}
	core_gonna_use  := count / thread_per_core
	for socket_index := 0; socket_index < len(sockets); socket_index++ {
		socket_id := sockets[socket_index]
		if socket_id != fmt.Sprintf("%d",numa_node) {
			continue
		}

		cores := make([]string, 0, len(all[socket_id]))
		for k := range all[socket_id] { cores = append(cores, k) }

		for core_index := 0; core_index < core_gonna_use; core_index++ {
			core_id   := cores[core_index]


			for thread := 0; thread < thread_per_core; thread++ {
				i,_ := strconv.ParseInt(all[socket_id][core_id][thread], 10, 32)

				vcpupin = append(vcpupin, struct{
					Vcpu   int "xml:\"vcpu,attr\""; 
					Cpuset int "xml:\"cpuset,attr\""
				}{
					Vcpu  : thread_per_core*core_index + thread,
					Cpuset: int(i),
				})
			}
		}
	}



	return vcpupin,nil
}