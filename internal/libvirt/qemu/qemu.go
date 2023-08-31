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

func newConn() (net.Conn, error) {
	return net.DialTimeout("unix", "/var/run/libvirt/libvirt-sock", 2*time.Second)
}
func NewQEMUHypervisor() *QEMUHypervisor {
	return &QEMUHypervisor{
		hypervisor: hypervisor.New(hypervisor.NewRPCDriver(newConn)),
	}
}

type Domain struct {
	Name   string      `json:"name"`
	Status qemu.Status `json:"status"`
}

func (qm *QEMUHypervisor) ListDomainWithStatus() []Domain {
	domains, err := qm.hypervisor.Domains()
	if err != nil {
		log.Fatalf("Unable to get domains from hypervisor: %v", err)
	}

	doms := []Domain{}
	for _, d := range domains {
		status,err := d.Status()
		if err != nil {
			continue
		}

		doms = append(doms, Domain{
			Status: status,
			Name:   d.Name,
		})
	}

	return doms
}
