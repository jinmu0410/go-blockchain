package test

import (
	"fmt"

	"github.com/jinmu/go-blockchain/internal/blockchain/ethereum"
	"testing"
)

func TestBalance(t *testing.T){
	balance, err := ethereum.Balance(ethereum.GetClient(), "0x71c7656ec7ab88b098defb751b7401b5f6d8976f")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(balance)
}
