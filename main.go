package main

import (
	"github.com/sah4ez/golang-test-task/Service"
)

func main() {
	httpService := Service.Server{}
	httpService.Listen()
}
