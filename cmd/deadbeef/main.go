// Read words and find all combinations that can be made using hex representation
package main

import (
	"deadbeef"
	"fmt"
	"github.com/mitchellh/go-linereader"
	"os"
	"strconv"
)

func main() {
	length, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic("%s is not a number")
	}

	outbound := linereader.New(os.Stdin).Ch
	inbound := deadbeef.Graph(outbound, length)

	for in := range inbound {
		fmt.Println(in)
	}
}
