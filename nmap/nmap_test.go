package nmap

import (
	"fmt"
	"testing"
)

func TestNmap(t *testing.T) {
	result := FindIPMac("192.168.1.*")
	fmt.Printf("%v\n", result)

}