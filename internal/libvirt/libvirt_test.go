package libvirt

import (
	"io"
	"os"
	"test/model"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestCreateVM(t *testing.T) {
	lv := NewLibvirt()

	file,err := os.OpenFile("/home/huyhoang/go-libvirt/test/model/data/vm7.yaml",os.O_RDWR,0755)
	if err != nil {
		panic(err)
	}

	data,err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	dom := model.Domain{}
	gpu := lv.ListGPUs()[0]
	yaml.Unmarshal(data, &dom)
	file.Close()

	
	_,_ = lv.CreateVM(dom,[]model.GPU{gpu})
}