package main

import (
	"fmt"
	"log"
	"os"

	"jamel/pkg/cve"
)

func main() {
	out, err := cve.Get(os.Args[1])
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(string(out))
}
