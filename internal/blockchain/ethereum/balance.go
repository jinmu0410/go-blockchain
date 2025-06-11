package ethereum

import (
	"context"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

/**
* Get the balance of an address
* @param client *Client
* @param address string
* @return string, error
*/
func Balance(client *Client, address string) (string, error) {
	hexAddress := common.HexToAddress(address)
	balance, err := client.Client.BalanceAt(context.Background(), hexAddress, nil)
	if err != nil {
		log.Println("Failed to get balance", err)
		return "", err
	}
	return balance.String(), nil
}

/**
* Get the balance of an address at a specific block number
* @param client *Client
* @param address string
* @param blockNumber uint64
* @return string, error
*/
func BalanceOfBlockNumber(client *Client, address string, blockNumber uint64) (string, error) {
	hexAddress := common.HexToAddress(address)
	number := big.NewInt(int64(blockNumber))
	balance, err := client.Client.BalanceAt(context.Background(), hexAddress, number)
	if err != nil {
		log.Println("Failed to get balance", err)
		return "", err
	}
	return balance.String(), nil
}

/**
* Get the pending balance of an address
* @param client *Client
* @param address string
* @return string, error
*/
func PendingBalance(client *Client, address string) (string, error) {
	hexAddress := common.HexToAddress(address)
	balance, err := client.Client.PendingBalanceAt(context.Background(), hexAddress)
	if err != nil {
		log.Println("Failed to get pending balance", err)
		return "", err
	}
	return balance.String(), nil
}
