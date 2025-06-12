package test

import (
	"testing"

	"github.com/jinmu/go-blockchain/internal/blockchain/ethereum"
)

func TestGetTransaction(t *testing.T) {
	ethereum.GetTransaction(ethereum.GetClient(), "0x5d49fcaa394c97ec8a9c3e7bd9e8388d420fb050a52083ca52ff24b3b65bc9c2")
}