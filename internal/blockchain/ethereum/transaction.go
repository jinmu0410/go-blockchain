package ethereum

import (
	"context"
	"fmt"
	"math/big"
	"log"
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

/**
* Get the block number
* @param client *Client
* @param blockNumber uint64
* @return uint64, error
 */
func GetBlockNumber(client *Client, blockNumber uint64) (uint64, error) {
	block, err := client.Client.BlockByNumber(context.Background(), big.NewInt(int64(blockNumber)))
	if err != nil {
		return 0, err
	}
	fmt.Println(block.Number().Uint64())
	fmt.Println(block.Hash().Hex())
	fmt.Println(block.Time())
	fmt.Println(block.Difficulty().Uint64())
	fmt.Println(block.Nonce())
	fmt.Println(block.Transactions().Len())

	return block.Number().Uint64(), nil
}

/**
* Get the transaction
* @param client *Client
* @param transactionHash string
* @return string, error
*/
func GetTransaction(client *Client, transactionHash string) (string, error) {
	hash := common.HexToHash(transactionHash)
	receipt, err := client.Client.TransactionReceipt(context.Background(), hash)
	if err != nil {
		return "", err
	}

	tx, _, err := client.Client.TransactionByHash(context.Background(), hash)
	if err != nil {
		return "", err
	}

	from, err := client.Client.TransactionSender(context.Background(), tx, receipt.BlockHash, receipt.TransactionIndex)
	if err != nil {
		return "", err
	}

	println(receipt.TxHash.Hex())
	println(receipt.Logs)
	println(receipt.Status)
	println(receipt.GasUsed)
	println(receipt.EffectiveGasPrice)
	println(receipt.BlockHash.Hex())
	println(receipt.BlockNumber.Uint64())
	println(receipt.TransactionIndex)
	println(from.Hex())
	println(tx.To().Hex())

	return receipt.TxHash.Hex(), nil
}

/**
* Transfer
* @param client *Client
* @param toAddress string
* @param amount string
* @return int, error
*/
func Transfer(client *Client, toAddress string, amount int) (string, error) {
	//private key
	privateKey, err := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19")
	if err != nil {
		return "", err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
    if !ok {
        log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
    }

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := client.Client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return "", err
	}

	value := big.NewInt(int64(amount)) // in wei (1 eth)
    gasLimit := uint64(21000)
	chainID, err := client.Client.ChainID(context.Background())
	if err != nil {
		return "", err
	}                // in units
    // gasPrice, err := client.Client.SuggestGasPrice(context.Background())
    // if err != nil {
    //     return "", err
    // }

	// tx := types.NewTransaction(nonce, common.HexToAddress(toAddress), value, gasLimit, gasPrice, nil)
	// 获取 baseFee (需支持 EIP-1559 的节点)
	header, err := client.Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return "", fmt.Errorf("get latest block header failed: %w", err)
	}
	baseFee := header.BaseFee

	// 设置 MaxPriorityFee 和 MaxFeePerGas（MaxFee >= BaseFee + Priority）
	maxPriorityFee := big.NewInt(2e9) // 2 Gwei
	maxFee := new(big.Int).Add(baseFee, maxPriorityFee)
	to :=common.HexToAddress(toAddress)

	// 构造 EIP-1559 类型交易
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: maxPriorityFee, // priority fee（小费）
		GasFeeCap: maxFee,         // max total fee
		Gas:       gasLimit,
		To:        &to,	
		Value:     value,
		Data:      nil,
	})

	signer := types.NewEIP155Signer(chainID)
	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		return "", err
	}

	err = client.Client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", err
	}

	txHash := signedTx.Hash().Hex()

	return txHash, nil
}


/**
* Subscribe to new block
* @param client *Client
* @return nil
*/
func SubscribeNewBlock(client *Client) {
	headers := make(chan *types.Header)
	sub, err := client.Client.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
			case err := <-sub.Err():
				log.Fatal(err)
			case header := <-headers:
				fmt.Println("New block detected, block number: ", header.Number.Uint64())
		}
	}
}

