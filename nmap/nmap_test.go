package nmap

import (
	"fmt"
	"testing"
)

func TestNmap(t *testing.T) {
	result := FindIPMac()
	fmt.Printf("%v\n", result)
}