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
	// 使用 Ankr 的公共节点，它通常更稳定
	return NewClient("https://eth.llamarpc.com")
}
