package ethereum

import (
	"log"

	"github.com/ethereum/go-ethereum/ethclient"
)

type Client struct {
	Client *ethclient.Client
}

func NewClient(url string) *Client {
	client, err := ethclient.Dial(url)
	if err != nil {
		log.Println("Failed to connect to Ethereum client", err)
		return nil
	}
	return &Client{Client: client}
}

func GetClient() *Client {
	return NewClient("https://mainnet.infura.io/v3/9aa3d95b3bc440fa88ea12eaa4456161")
}