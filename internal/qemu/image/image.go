package qemuimg

import (
	"fmt"
	"os/exec"
)

func CloneVolume(src string,
				 dst string, 
				 size int,
				 ) (error) {
	cmd := exec.Command("qemu-img","create",
		"-f", "qcow2" ,
		"-o", fmt.Sprintf("backing_file=%s", src),
		dst,
		fmt.Sprintf("%dG", size),
	)

	fmt.Printf("cloning disk with command %v\n",cmd.Args)

	out,err := cmd.Output()
	if err != nil {
		return fmt.Errorf("%s : %s",err.Error(),out) 
	}

	fmt.Printf("clone img done, result: %s\n",string(out))
	return nil
}
