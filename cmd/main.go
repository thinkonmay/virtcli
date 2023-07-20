package main

import (
	"fmt"
	"os"
	virtdaemon "test"

	"gopkg.in/yaml.v3"
)

func main() {
	result,err := virtdaemon.NewVirtDaemon(os.Args[1],[]byte(""))
	if err != nil {
		return
	}

	data,err := yaml.Marshal(result)
	if err != nil {
		return
	}

	fmt.Printf("%v\n", string(data))
}