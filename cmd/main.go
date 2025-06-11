package main

import (
	"fmt"
	"github.com/jinmu/go-blockchain/internal/blockchain/ethereum"
)

func main() {
	fmt.Println("Hello, World!")

	client := ethereum.GetClient()

	fmt.Printf("client: %v\n", client)

	
}