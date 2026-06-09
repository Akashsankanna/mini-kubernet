package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func main() {

	bytes := make([]byte, 32)

	_, err := rand.Read(bytes)

	if err != nil {
		panic(err)
	}

	fmt.Println(hex.EncodeToString(bytes))
}
