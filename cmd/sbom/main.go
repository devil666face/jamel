package main

import (
	"fmt"
	"log"
	"os"

	"jamel/pkg/sbom"
)

func main() {
	out, err := sbom.Get(os.Args[0])
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(string(out))
}
