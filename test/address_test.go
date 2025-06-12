package test

import (
	"github.com/jinmu/go-blockchain/internal/blockchain/ethereum"
	"testing"
)

func TestCreateAddress(t *testing.T){
	ethereum.CreateAddress(ethereum.GetClient())
}
