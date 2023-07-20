package ovs

import (
	"fmt"
	"testing"
)

func TestOVS(t *testing.T) {
	sw := NewOVS()
	brs,err := sw.Status()
	if err != nil {
		t.Error(err.Error())
	}
	fmt.Printf("%s\n", brs)
}