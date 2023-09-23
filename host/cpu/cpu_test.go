package cpu

import (
	"testing"
	"fmt"
)

func TestCPU(t *testing.T) {
	out,_ := GetHostTopology()
	fmt.Printf("%s\n",out)
}