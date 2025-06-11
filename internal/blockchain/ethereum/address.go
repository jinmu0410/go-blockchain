package ethereum

import (
	"context"
	"log"
	"regexp"

	"github.com/ethereum/go-ethereum/common"
)

/**
* Check if an address is valid
* @param address string
* @return bool
*/
func CheckIsValid(address string) bool {
	regex := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
	return regex.MatchString(address)
}

/**
* Check if an address is a contract
* @param client *Client
* @param address string
* @return bool
*/
func CheckAddressIsContract(client *Client, address string) bool {
	hexAddress := common.HexToAddress(address)
	byCode, err := client.Client.CodeAt(context.Background(), hexAddress, nil)
	if err != nil {
		log.Println("Failed to get code", err)
	}
	return len(byCode) > 0
}