package ovs

import (
	"fmt"
	"test/model"
	"time"

	"github.com/digitalocean/go-openvswitch/ovs"
)

type OVSPort struct {
	Name string `json:"name"`
}
type OVSBridge struct {
	Name string `json:"name"`
	Ports []OVSPort `json:"ports"`
}

type OVSState struct {

}

type OpenVSwitch struct {
	svc *ovs.Client
}

func NewOVS() *OpenVSwitch {
	svc := ovs.New()
	svc.VSwitch.AddBridge("br")
	svc.VSwitch.AddPort("br","enp5s0")
	return &OpenVSwitch{
		svc: svc,
	}
}

func (ovs *OpenVSwitch) Status() (*[]OVSBridge,error){
	ret := []OVSBridge{}
	brs,err := ovs.svc.VSwitch.ListBridges()
	for _, v := range brs {
		br := OVSBridge{Name: v}
		ports,err := ovs.svc.VSwitch.ListPorts(br.Name)
		if err != nil {
			return nil, err
		}

		for _, v2 := range ports {
			br.Ports = append(br.Ports, OVSPort{
				Name: v2,
			})
		}

		ret = append(ret, br)
	}

	return &ret,err
}




func (ovs *OpenVSwitch) CreateInterface() (*model.Interface,error) {
	now := fmt.Sprintf("%d",time.Now().UnixMilli())
	err := ovs.svc.VSwitch.AddPort("br",now)
	if err != nil {
		return nil, err
	}

	bridge,Type := "br","e1000e"
	// AddrType:= "pci"
	// Domain:= "0x0000"
	// Bus:= "0x01"
	// Slot:= "0x00"
	// Function:= "0x0"
	return &model.Interface{
		Type: "bridge",
		VirtualPort: &struct{Type string "xml:\"type,attr\""}{
			Type: "openvswitch",
		},
		Source: &model.InterfaceSource{
			Bridge: &bridge,
		},
		Target: &struct{Dev string "xml:\"dev,attr\""}{
			Dev: now,
		},
		Model: &struct{Type *string "xml:\"type,attr\""}{
			Type: &Type,
		},
		// Address: &model.InterfaceAddr{
		// 	Type: &AddrType,
		// 	Domain: &Domain,
		// 	Bus: &Bus,
		// 	Slot: &Slot,
		// 	Function: &Function,
		// },
	},nil
}




func (ovs *OpenVSwitch) DestroyOVS() error {
	status,err := ovs.Status()
	if err != nil {
		return err
	}

	for _,o := range (*status) {
		for _,o2 := range o.Ports {
			err := ovs.svc.VSwitch.DeletePort(o.Name,o2.Name)
			if err != nil {
				return err
			}

		}
		err := ovs.svc.VSwitch.DeleteBridge(o.Name)
		if err != nil {
			return err
		}
	}

	return nil
}