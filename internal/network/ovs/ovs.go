package ovs

import (
	"fmt"
	"net"
	"test/internal/network"
	"test/internal/network/ovs/arp"
	"test/model"
	"time"

	"github.com/digitalocean/go-openvswitch/ovs"
)

type OpenVSwitch struct {
	svc *ovs.Client
}

func NewOVS(iface string) network.Network {
	ifis,_ := net.Interfaces()
	throw := true
	for _, i2 := range ifis {
		if i2.Name == iface {
			throw = false
		}
	}
	if throw {
		panic(fmt.Errorf("not iface was found %s",iface))
	}

	svc := ovs.New()
	svc.VSwitch.AddBridge("br")
	svc.VSwitch.AddPort("br", iface)
	return &OpenVSwitch{
		svc: svc,
	}
}

func (ovs *OpenVSwitch) CreateInterface() (*model.Interface, error) {
	now := fmt.Sprintf("%d", time.Now().UnixMilli())
	err := ovs.svc.VSwitch.AddPort("br", now)
	if err != nil {
		return nil, err
	}

	bridge, Type := "br", "e1000e"
	return &model.Interface{
		Type: "bridge",
		VirtualPort: &struct {
			Type string "xml:\"type,attr\""
		}{
			Type: "openvswitch",
		},
		Source: &model.InterfaceSource{
			Bridge: &bridge,
		},
		Target: &struct {
			Dev string "xml:\"dev,attr\""
		}{
			Dev: now,
		},
		Model: &struct {
			Type *string "xml:\"type,attr\""
		}{
			Type: &Type,
		},
	}, nil
}

// FindDomainIPs implements network.Network.
func (*OpenVSwitch) FindDomainIPs(dom model.Domain) []string {
	return arp.FindDomainIPs(dom)
}

