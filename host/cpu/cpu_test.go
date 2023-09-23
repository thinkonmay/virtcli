package cpu

import (
	"testing"
	"fmt"
	"sort"
	"strconv"
)

func TestCPU(t *testing.T) {
	available,_ := GetHostTopology()

	all := map[string]map[string][]string{}
	for _,core := range available.CPUs {
		if all[core.Socket] == nil {
			all[core.Socket] = map[string][]string{ core.Core : {} }
		}

		all[core.Socket][core.Core] = append(all[core.Socket][core.Core], core.CPU)
	}

	sockets := make([]string, 0, len(all))
	for k := range all { sockets = append(sockets, k) }
	
	cores := make([]string, 0, len(all[sockets[1]]))
	for k := range all[sockets[0]] { cores = append(cores, k) }

	sort.Slice(cores, func(i, j int) bool {
		a,_ := strconv.ParseInt(cores[i],10,32)
		b,_ := strconv.ParseInt(cores[j],10,32)

		return a < b
	})

	fmt.Printf("%v",cores)
}