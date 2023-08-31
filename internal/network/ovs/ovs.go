package ovs

import (
	"fmt"
	"test/internal/network"
	"test/internal/network/ovs/arp"
	"test/model"
	"time"

	"github.com/digitalocean/go-openvswitch/ovs"
)

type OpenVSwitch struct {
	svc *ovs.Client
}

func NewOVS() network.Network {
	svc := ovs.New()
	svc.VSwitch.AddBridge("br")
	svc.VSwitch.AddPort("br", "enp5s0")
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

