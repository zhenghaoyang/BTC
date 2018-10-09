package main

import (
	"log"
)

func main() {
	log.Println("mining btc starting")

	cli := CLI{}

	cli.Run()

}
