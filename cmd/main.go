package main

import (
	"encoding/base64"
	"fmt"
	"os"
	virtdaemon "test"

	"gopkg.in/yaml.v3"
)

func main() {
	sec := []byte("")
	if len(os.Args) == 3 {
		var err error
		sec ,err = base64.StdEncoding.DecodeString(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to decode second argument: %s", err.Error())
			os.Exit(2)
		}
	} else if len(os.Args) == 1 {
		fmt.Fprintf(os.Stderr, "not enough number of argument")
		os.Exit(2)
	}

	result,err := virtdaemon.NewVirtDaemon(os.Args[1],sec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "task failed: %s",err.Error())
		return
	}

	data,err := yaml.Marshal(result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to unmarshal output %s",err.Error())
		return
	}

	fmt.Printf("%v\n", string(data))
	os.Exit(55)
}