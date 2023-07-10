package libvirt

import (
	"fmt"
	"test/model"
	"testing"
)

func TestCreateVM(t *testing.T) {
	lv := NewLibvirt()

	gpu 	:= lv.ListGPUs()
	vols 	:= lv.ListDisks()

	vol := []model.Volume{}
	for _, v := range vols {
		if v.Path == "/disk/2TB1/AtlatOS-copy.qcow2" ||
		   v.Path == "/disk/2TB1/AtlasOS-cloned.qcow2" {
			
			vol = append(vol, v)
		}
	}

	_,err := lv.CreateVM(8,16,gpu[:1],vol[1:2])
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}

	
	_,err = lv.CreateVM(10,12,gpu[1:],vol[:1])
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}


}
func TestBackingChain(t *testing.T) {
	lv := NewLibvirt()
	vols := lv.ListDisks()
	for _, v := range vols {
		if v.Name == "990035145.qcow2" {
			fmt.Printf("%s",backingChain(vols,v).ToString())
			return
		}
	}
}

func TestIP(t *testing.T) {
	lv := NewLibvirt()
	vols := lv.ListDomains()

	for _, d := range vols {
		if *d.Name == "vm1" {
			result := lv.ListDomainIPs(d)
			fmt.Println(result)
		}
	}
}