package main

import virtdaemon "test"

func main() {
	virtdaemon.NewVirtDaemon(8090)
	<-make(chan bool)
}