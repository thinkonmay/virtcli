package libvirt

import (
	"fmt"
	"io"
	"os"
	"test/model"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestCreateVM(t *testing.T) {
	lv := NewLibvirt()

	file,err := os.OpenFile("../../model/data/vm.yaml",os.O_RDWR,0755)
	if err != nil {
		panic(err)
	}

	data,err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	dom := model.Domain{}
	gpu := lv.ListGPUs()
	vols := lv.ListDisks()
	ifs := lv.ListIfaces()
	yaml.Unmarshal(data, &dom)
	file.Close()

	vol := []model.Volume{}
	for _, v := range vols {
		if v.Path == "/disk/2TB1/AtlatOS-copy.qcow2" ||
		   v.Path == "/disk/2TB1/AtlasOS-cloned.qcow2" {
			
			vol = append(vol, v)
		}
	}

	dom.Input = []model.Input{}
	
	_,err = lv.CreateVM(dom,gpu[:1],vol[1:2],ifs[1:2])
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}

	
	_,err = lv.CreateVM(dom,gpu[1:],vol[:1],ifs[:1])
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}
}