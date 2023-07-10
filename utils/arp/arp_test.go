package arp

import (
	"fmt"
	"testing"
)

func TestArp(t *testing.T) {
	macip,err := getIPIface()
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("%s\n", macip)
}