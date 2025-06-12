package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jinmu/go-blockchain/internal/blockchain/ethereum"
)

func main() {
	client := ethereum.GetClient()
	if client == nil {
		log.Fatal("Failed to connect to Ethereum node")
	}

	// 测试连接
	_, err := client.Client.BlockNumber(context.Background())
	if err != nil {
		log.Fatalf("Failed to get block number: %v", err)
	}

	fmt.Println("Successfully connected to Ethereum node")

	// 测试获取余额
	balance, err := ethereum.Balance(client, "0x71c7656ec7ab88b098defb751b7401b5f6d8976f")
	if err != nil {
		log.Fatalf("Failed to get balance: %v", err)
	}
	fmt.Printf("Balance: %s\n", balance)
}
