package qemuimg

import (
	"fmt"
	"os/exec"
)

func CloneVolume(src string,
				 dst string, 
				 size int,
				 ) (error) {
	out,err := exec.Command("qemu-image","create","-o", 
		fmt.Sprintf("backing_file=%s", src),
		dst,
		fmt.Sprintf("%dG", size),
	).Output()
	if err != nil {
		return err
	}

	fmt.Printf("clone img done, result: %s\n",string(out))
	return nil
}
