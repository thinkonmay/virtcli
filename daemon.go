package virtdaemon

import (
	"fmt"
	"test/internal/libvirt"
	qemuhypervisor "test/internal/qemu"
	"test/model"
	"time"

	"github.com/digitalocean/go-qemu/qemu"
	"gopkg.in/yaml.v3"
)



type VirtDaemon struct {
	hypervisor *qemuhypervisor.QEMUHypervisor
	libvirt *libvirt.Libvirt
}

func NewVirtDaemon(verb string, data []byte) (any,error){
	daemon := &VirtDaemon{
		hypervisor: qemuhypervisor.NewQEMUHypervisor(),
		libvirt: libvirt.NewLibvirt(),
	}

	var fun func(data []byte)(any,error)
	switch verb {
	case "/deploy": 		
		fun = daemon.deployVM
	case "/start": 		
		fun = daemon.startVM
	case "/stop": 		
		fun = daemon.stopVM
	case "/delete": 		
		fun = daemon.deleteVM
	case "/status": 		
		fun = daemon.statusVM
	case "/vms": 		
		fun = daemon.listVMs
	case "/gpus": 		
		fun = daemon.listGPUs
	}

	return fun(data)
}




func backingChain(vols *Volume) *model.BackingStore {
	if vols == nil {
		return nil
	}

	return &model.BackingStore{
		Type: "file",
		Format: &struct{Type string "xml:\"type,attr\""}{
			Type: "qcow2",
		},
		Source: &struct{File string "xml:\"file,attr\""}{
			File: vols.Path,
		},
		BackingStore: backingChain(vols.Backing),
	}
}


type Volume struct{
	Path string `yaml:"path"`
	Backing *Volume `yaml:"backing"`
}

func (daemon *VirtDaemon)deployVM(body []byte) (any, error) {
	server := struct{
		ID   string `yaml:"id"`
		VCPU int `yaml:"vcpus"`
		RAM  int `yaml:"ram"`
		GPU []model.GPU `yaml:"gpu"`
		Volumes []Volume `yaml:"volumes"`
	}{}

	err := yaml.Unmarshal(body,&server)
	if err != nil {
		return nil,err
	}



	volumes := []model.Disk{}
	for i,v := range server.Volumes {
		dev := ""
		switch i {
		case 0:
		dev = "hda"
		case 1:
		dev = "hdb"
		case 2:
		dev = "hdc"
		case 3:
		dev = "hdd"
		case 4:
		dev = "hde"
		case 5:
		dev = "hdf"
		}


		volumes = append(volumes, model.Disk{
			Driver: &struct{Name string "xml:\"name,attr\""; Type string "xml:\"type,attr\""}{
				Name: "qemu",
				Type: "qcow2",
			},
			Source: &struct{File string "xml:\"file,attr\""; Index int "xml:\"index,attr\""}{
				File: v.Path,
				Index: 1,
			},
			Target: &struct{Dev string "xml:\"dev,attr\""; Bus string "xml:\"bus,attr\""}{
				Dev: dev,
				Bus: "ide",
			},
			Address: nil,
			Type: "file",
			Device: "disk",
			BackingStore: backingChain(v.Backing),
		})	
	}

	name,err := daemon.libvirt.CreateVM(
		server.ID,
		server.VCPU,
		server.RAM,
		server.GPU,
		volumes,
	)

	return struct {
		Name string
	} {
		Name: name,
	},err
}

func (daemon *VirtDaemon)stopVM(body []byte) (any, error) {
	server := struct{
		Name string `yaml:"name"`
	}{}

	err := yaml.Unmarshal(body,&server)
	if err != nil {
		return nil,err
	}



	start := time.Now()
	for {
		if time.Now().UnixMilli() - start.UnixMilli() > 30 * 1000 {
			break
		}

		doms := daemon.hypervisor.ListDomain()
		for _, d := range doms {
			if d.Name == string(server.Name) {
				if d.Status == qemu.StatusRunning {
					err = daemon.libvirt.StopVM(server.Name)
					if err != nil {
						return nil, err
					}
				} else if d.Status == qemu.StatusShutdown {
					return fmt.Sprintf("VM %s stopped",server.Name),nil
				} else {
					return nil,fmt.Errorf("VM %s in %s state, abort",server.Name,d.Status.String())
				}
			}
		}

		time.Sleep(time.Second)
	}

	return nil,fmt.Errorf("timeout shuting down VM %s",server.Name)
}


func (daemon *VirtDaemon)startVM(body []byte) (any, error) {
	server := struct{
		Name string `yaml:"name"`
		GPU []model.GPU `yaml:"gpu"`
	}{}

	err := yaml.Unmarshal(body,&server)
	if err != nil {
		return nil,err
	}


	found := false
	doms := daemon.hypervisor.ListDomain()
	for _, d := range doms {
		if d.Name == string(server.Name) && 
		   d.Status == qemu.StatusShutdown {
			found = true
		}
	}
	if !found {
		return nil,fmt.Errorf("vm %s not found",string(server.Name))
	}


	err = daemon.libvirt.StartVM(server.Name,server.GPU)
	return fmt.Sprintf("VM %s stopped",server.Name),err
}


func (daemon *VirtDaemon)deleteVM(body []byte) (any, error) {
	server := struct{
		Name string `yaml:"name"`
	}{}

	err := yaml.Unmarshal(body,&server)
	if err != nil {
		return nil,err
	}


	found := false
	running := false
	doms := daemon.hypervisor.ListDomain()
	for _, d := range doms {
		if d.Name == string(server.Name) {
			found = true
			running = d.Status == qemu.StatusRunning
		}
	}
	if !found {
		return nil,fmt.Errorf("vm %s not found",string(server.Name))
	}


	err = daemon.libvirt.DeleteVM(server.Name,running)
	return fmt.Sprintf("VM %s deleted",server.Name),err
}


func (daemon *VirtDaemon)statusVM(body []byte) (any, error) {
	server := struct{
		Name string `yaml:"name"`
	}{}

	err := yaml.Unmarshal(body,&server)
	if err != nil {
		return nil,err
	}

	doms := daemon.hypervisor.ListDomain()
	for _, d := range doms {
		if d.Name == string(server.Name) {
			return struct{
				Status string
			}{
				Status: d.Status.String(),
			},nil
		}
	}

	return struct{
		Status string
	}{
		Status: "StatusDeleted",
	},nil
}


func (daemon *VirtDaemon)listVMs(data []byte) (any, error) {
	doms    := daemon.libvirt.ListDomains()
	qemudom := daemon.hypervisor.ListDomain()

	result := map[string][]model.Domain{}

	for _, d := range qemudom {
		for _, d2 := range doms {
			if d.Name == *d2.Name {
				if result[d.Status.String()] == nil {
					result[d.Status.String()] = []model.Domain{d2}
				} else {
					result[d.Status.String()] = append(result[d.Status.String()],d2)
				}
			}
		}
	}

	for i, d := range result["StatusRunning"] {
		ips := daemon.libvirt.ListDomainIPs(d)
		result["StatusRunning"][i].PrivateIP = &ips
	}

	return result,nil
}







func (daemon *VirtDaemon)listGPUs(data []byte) (any,error) {
	gpus 	:= daemon.libvirt.ListGPUs()

	domains := daemon.libvirt.ListDomains()
	qemudom := daemon.hypervisor.ListDomain()

	result := struct{
		Active []model.GPU `yaml:"active"`
		Available []model.GPU `yaml:"open"`
	}{
		Active: []model.GPU{},
		Available: []model.GPU{},
	}

	for _, g := range gpus {
		add := true
		for _, d := range domains {

			ignore := false
			for _, d2 := range qemudom {
				if *d.Name == d2.Name && 
					d2.Status != qemu.StatusRunning &&
					d2.Status != qemu.StatusPaused {
					ignore = true
				}
			}
			
			if ignore {
				continue
			}


			for _, hd := range d.Hostdevs {
				for _, v := range g.Capability.IommuGroup.Address {
					if  hd.SourceAddress.Bus == v.Bus &&
						hd.SourceAddress.Domain == v.Domain && 
						hd.SourceAddress.Function == v.Function && 
						hd.SourceAddress.Slot == v.Slot {
						g.VM = d.Name
						add = false
					}
				}
			}
		}

		if add {
			result.Available = append(result.Available, g)
		}
		result.Active = append(result.Active, g)
	}


	return result,nil
}