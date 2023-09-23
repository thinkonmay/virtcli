package cpu

import (
	"os/exec"
	"encoding/json"
)

type HostCore struct {
	CPU int
	Node int
	Socket int
	Core int
	Cache string `json:"l1d:l1i:l2:l3"`
}

type LsCPU struct {
	CPUs []HostCore `json:"cpus"`
}

func GetHostTopology() (*LsCPU,error){
	out,err := exec.Command("lscpu","-J","-e").Output()
	if err != nil {
		return nil,err
	}



	ret := &LsCPU{}
	err = json.Unmarshal(out,ret)
	if err != nil {
		return nil,err
	}

	return ret,nil
}


