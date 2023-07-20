// Copyright 2016 The go-qemu Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package qemuhypervisor

import (
	"log"
	"net"
	"time"

	hypervisor "github.com/digitalocean/go-qemu/hypervisor"
	"github.com/digitalocean/go-qemu/qemu"
)


type QEMUHypervisor struct {
	hypervisor *hypervisor.Hypervisor

}

func NewQEMUHypervisor()  *QEMUHypervisor {
	qmhv := &QEMUHypervisor{}

	network := "unix"
	address := "/var/run/libvirt/libvirt-sock"
	timeout := 2*time.Second
	newConn := func() (net.Conn, error) {
		return net.DialTimeout(network, address, timeout)
	}

	driver := hypervisor.NewRPCDriver(newConn)
	qmhv.hypervisor = hypervisor.New(driver)

	return qmhv
}



type Domain struct {
	Name 		string   			`json:"name"`	
	Status 		qemu.Status			`json:"status"`	

	PcieDevs 	[]qemu.PCIDevice	`json:"pcies"`	
	BlockDevs 	[]qemu.BlockDevice	`json:"blocks"`	
	CPUs 		[]qemu.CPU			`json:"cpus"`	
}

func (qm *QEMUHypervisor)ListDomain() []Domain {
	domains, err := qm.hypervisor.Domains()
	if err != nil {
		log.Fatalf("Unable to get domains from hypervisor: %v", err)
	}

	doms := []Domain{}
	for _, d := range domains {
		dom := Domain{}
		
		if status,err := d.Status(); status != qemu.StatusRunning || err != nil{
			doms = append(doms, Domain{
				Status: status,
				Name: d.Name,
			})
			continue
		} else {
			dom.Status = status
			dom.Name   = d.Name
		}


		dom.PcieDevs,err = d.PCIDevices()
		if err != nil {
			log.Fatalf("Unable to get domains from hypervisor: %v", err)
			continue
		}

		dom.BlockDevs,err = d.BlockDevices()
		if err != nil {
			log.Fatalf("Unable to get domains from hypervisor: %v", err)
			continue
		}

		dom.CPUs,err = d.CPUs()
		if err != nil {
			log.Fatalf("Unable to get domains from hypervisor: %v", err)
			continue
		}

		doms = append(doms, dom)
	}

	return doms
}