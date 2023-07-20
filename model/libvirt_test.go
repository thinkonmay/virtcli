package model 

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func convertToYaml(data []byte) []byte {
	v := Domain{}
	err := xml.Unmarshal(data, &v)
	if err != nil {
		fmt.Printf("error: %v", err)
		return []byte("")
	}

	data,_= yaml.Marshal(v)
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

	file, err := os.OpenFile("./data/vm7.xml", os.O_RDWR, 0755)
	if err != nil {
		panic(err)
	}

	data, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	file.Close()

	encoded := convertToYaml(data)
	file, err = os.OpenFile("./data/vm7.yaml", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}
	file.Truncate(0)
	file.WriteAt(encoded, 0)
	file.Close()

	data = convertToXML(encoded)
	file, err = os.OpenFile("./data/vm7.out.xml", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}
	file.Truncate(0)
	file.WriteAt(data, 0)
	file.Close()
}