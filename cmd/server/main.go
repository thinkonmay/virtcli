package main

import (
	"fmt"
	virtdaemon "test"

	"gopkg.in/yaml.v3"
)

func main() {
	result,err := virtdaemon.NewVirtDaemon("/vms",[]byte(""))
	if err != nil {
		return
	}

	data,err := yaml.Marshal(result)
	if err != nil {
		return
	}

	fmt.Printf("%v\n", string(data))
}