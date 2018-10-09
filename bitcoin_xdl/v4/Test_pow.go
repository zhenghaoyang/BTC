package main

import (
	"crypto/sha256"
	"fmt"
	"strconv"
	"time"
)

func testmain() {
	start := time.Now()

	for i := 0; i < 1000000; i++ {
		data := sha256.Sum256([]byte(strconv.Itoa(i)))

		fmt.Printf("%10d , %x\n", i, data)
		fmt.Printf("末尾为 %v\n", string(data[len(data)-1:]))
		if string(data[len(data)-1:]) == "0" {
			usedtime := time.Since(start)
			fmt.Println("succee %d ms", usedtime)
			break
		}

	}
}
