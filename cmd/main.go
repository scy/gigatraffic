package main

import (
	"log"
	"github.com/scy/gigatraffic"
)

func main() {
	quota, err := gigatraffic.Retrieve()
	if err != nil {
		log.Fatal(err)
	}
	log.Print(quota.String())
}
