package main

import (
	"os"
	"github.com/sah4ez/golang-test-task/Service"
)

func main() {
	hostArg := os.Args[1]
	portArg := os.Args[2]
	httpService := Service.Server{Host: hostArg, Port: portArg }
	httpService.Listen()
}
