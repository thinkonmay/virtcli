package libvirt

import (
	"fmt"
	// "test/model"
	"testing"
)

// func TestCreateVM(t *testing.T) {
// 	lv := NewLibvirt()

// 	gpu 	:= lv.ListGPUs()

// 	vol := []model.Volume{}


// 	_,err := lv.CreateVM(8,16,gpu[:1],vol[1:2])
// 	if err != nil {
// 		fmt.Printf("%s\n", err.Error())
// 	}

	
// 	_,err = lv.CreateVM(10,12,gpu[1:],vol[:1])
// 	if err != nil {
// 		fmt.Printf("%s\n", err.Error())
// 	}


// }


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