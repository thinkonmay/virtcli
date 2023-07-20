package arp

import (
	"fmt"
	"testing"
)

func TestArp(t *testing.T) {
	macip,err := getIPIface("192.168.1")
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("%s\n", macip)
}