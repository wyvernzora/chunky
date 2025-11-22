package main

import (
	"fmt"
	"os"

	"github.com/wyvernzora/chunky-md/pkg/chunky"
)

func main() {
	name := ""
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	fmt.Println(chunky.Hello(name))
}
