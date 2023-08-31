package libvirtnetwork

import (
	"fmt"
	"log"
	"net"
	"test/internal/network"
	"test/model"
	"time"

	"github.com/digitalocean/go-libvirt"
	"github.com/digitalocean/go-libvirt/socket/dialers"
)

const (
	network_name = "virtless-default-network"
	bridge_name  = "virtless-default-bridge"
)

func newNetwork(card string) string {
	return fmt.Sprintf(`
	<network>
		<name>%s</name>
		<forward dev="%s" mode="nat">
			<interface dev="%s"/>
			<nat>
				<port start="1024" end="65535"/>
			</nat>
		</forward>
		<bridge name="%s" stp="on" delay="0"/>
		<domain name="network"/>
		<ip address="192.168.100.1" netmask="255.255.255.0">
			<dhcp>
				<range start="192.168.100.1" end="192.168.100.254"/>
			</dhcp>
		</ip>
	</network>
	`, network_name, card, card, bridge_name)
}

type LibvirtNetwork struct {
	conn *libvirt.Libvirt
}



func NewLibvirtNetwork() network.Network {
	c, err := net.DialTimeout("unix", "/var/run/libvirt/libvirt-sock", 2*time.Second)
	if err != nil {
		log.Fatalf("failed to dial libvirt: %v", err)
	}

	ret := &LibvirtNetwork{
		conn: libvirt.NewWithDialer(dialers.NewAlreadyConnected(c)),
	}

	if err := ret.conn.Connect(); err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	return ret
}

func (ovs *LibvirtNetwork) CreateInterface() (*model.Interface, error) {
	Type := "e1000e"
	return &model.Interface{
		Type: "network",
		Model: &struct {
			Type *string "xml:\"type,attr\""
		}{
			Type: &Type,
		},
	}, nil
}

func (ovs *LibvirtNetwork) getIPMac() (map[string]string,error) {
	nets, _, err := ovs.conn.ConnectListAllNetworks(1, libvirt.ConnectListNetworksActive)
	if err != nil {
		return map[string]string{},err
	}

	ipmacs := map[string]string{}
	for _, n := range nets {
		leases, _, err := ovs.conn.NetworkGetDhcpLeases(n, []string{}, 1, 0)
		if err != nil {
			panic(err)
		}
		for _, ndl := range leases {
			for _, v := range ndl.Mac {
				ipmacs[v] = ndl.Ipaddr
			}
		}
	}

	return ipmacs,nil
}

func (network *LibvirtNetwork) FindDomainIPs(dom model.Domain) []string {
	ips := []string{}
	macs := []string{}
	for _, i2 := range dom.Interfaces {
		macs = append(macs, *i2.Mac.Address)
	}

	database,err := network.getIPMac()
	if err != nil {
		return []string{}
	}

	for k, v := range database {
		for _, v2 := range macs {
			if v2 == k {
				ips = append(ips, v)
			}
		}
	}

	return ips
}