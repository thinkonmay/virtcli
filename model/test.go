package model 

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func convertToYaml() []byte {
	v := Domain{}

	file, err := os.OpenFile("./temp.xml", os.O_RDWR, 0755)
	if err != nil {
		panic(err)
	}

	data, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	file.Close()

	err = xml.Unmarshal(data, &v)
	if err != nil {
		fmt.Printf("error: %v", err)
		return []byte("")
	}

	data,_= yaml.Marshal(v)
	fmt.Println(string(data))
	return data
}

func convertToXML(data []byte) []byte {
	v := Domain{}

	err := yaml.Unmarshal(data, &v)
	if err != nil {
		fmt.Printf("error: %v", err)
		return []byte("")
	}

	data,_= xml.MarshalIndent(v,"","  ")
	return data
}

func TestModel(t *testing.T) {
	data := convertToXML(convertToYaml())
	file, err := os.OpenFile("./temp1.xml", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}
	file.Truncate(0)
	file.WriteAt(data, 0)
	file.Close()
}