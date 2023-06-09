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
	ifs 	:= lv.ListIfaces()

	vol := []model.Volume{}
	for _, v := range vols {
		if v.Path == "/disk/2TB1/AtlatOS-copy.qcow2" ||
		   v.Path == "/disk/2TB1/AtlasOS-cloned.qcow2" {
			
			vol = append(vol, v)
		}
	}

	i := []model.Iface{}
	for _, v := range ifs {
		if v.Name == "enp0s25" ||
		   v.Name == "enp5s0" {
			continue
		}
		i = append(i, v)
	}

	
	_,err := lv.CreateVM(8,16,gpu[:1],vol[1:2],i[1:2])
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}

	
	_,err = lv.CreateVM(10,12,gpu[1:],vol[:1],i[:1])
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}


}