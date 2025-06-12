package ethereum

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"crypto/ecdsa"
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

/**
* Create a new address
* @param client *Client
* @return string, error
*/
func CreateAddress(client *Client) (string, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}
	privateKeyBytes := crypto.FromECDSA(privateKey)
    fmt.Println("Private Key: ", hexutil.Encode(privateKeyBytes)[2:])

	publicKey := privateKey.Public()
	
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}

	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	fmt.Println(hexutil.Encode(publicKeyBytes)[4:])

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	fmt.Println("Address: ", address)

	return address, nil
}