package client

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"log"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Client struct {
	c *ethclient.Client
}

func New(chainURL string) *Client {
	client, err := ethclient.Dial(chainURL)
	if err != nil {
		log.Fatal(err)
	}
	return &Client{c: client}
}

func (c *Client) LatestBlock() (string, error) {
	header, err := c.c.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return "", err
	}

	return header.Number.String(), nil
}

// Last10Tx fetches last 10 transactions
// going from last block to the first one we collect all transactions(not clear how many tx in 1 block),
// so we stop either when 1st block riched or when we have collected 10 transactions
func (c *Client) Last10Tx() (string, error) {
	header, err := c.c.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return "", err
	}

	var bts bytes.Buffer
	b := header.Number
	bIntZero := big.NewInt(0)
	txCount := 0
	for {
		//block less than 0
		if b.Cmp(bIntZero) == -1 || txCount >= 10 {
			break
		}

		block, err := c.c.BlockByNumber(context.Background(), header.Number)
		if err != nil {
			return "", err
		}
		for _, tx := range block.Transactions() {
			txCount++
			if txCount > 10 {
				break
			}
			bts.WriteString("===\n")
			bts.WriteString(fmt.Sprintf("txHash: %s\n", tx.Hash().Hex()))
			bts.WriteString(fmt.Sprintf("txValue: %s\n", tx.Value().String()))
			fbalance := new(big.Float)
			fbalance.SetString(tx.Value().String())
			ethValue := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(18)))
			bts.WriteString(fmt.Sprintf("txValueEth: %f\n", ethValue))
			bts.WriteString(fmt.Sprintf("txGas: %d\n", tx.Gas()))
			bts.WriteString(fmt.Sprintf("GasPrice: %d\n", tx.GasPrice().Uint64()))
			bts.WriteString(fmt.Sprintf("txData: %s\n", tx.Data()))
			bts.WriteString(fmt.Sprintf("txTo: %s\n", tx.To().Hex()))
			bts.WriteString("===\n")
		}
		b = b.Sub(b, big.NewInt(1))
	}

	return bts.String(), nil
}

func (c *Client) Balance(acc string) (string, error) {
	account := common.HexToAddress(acc)
	balance, err := c.c.BalanceAt(context.Background(), account, nil)
	if err != nil {
		return "", err
	}

	var bts bytes.Buffer
	bts.WriteString("===\n")
	bts.WriteString(fmt.Sprintf("wallet: %s\n", acc))
	bts.WriteString(fmt.Sprintf("balance: %s\n", balance))
	fbalance := new(big.Float)
	fbalance.SetString(balance.String())
	ethValue := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(18)))
	bts.WriteString(fmt.Sprintf("txValueEth: %f\n", ethValue))
	bts.WriteString("===\n")

	return bts.String(), nil
}

// SendFunds sends funds from pkey account to acc account
func (c *Client) SendFunds(pkey, acc string, amount float64) (string, error) {
	privateKey, err := crypto.HexToECDSA(pkey)
	if err != nil {
		return "", fmt.Errorf("crypto.HexToECDSA(pkey): %w", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", errors.New("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := c.c.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return "", fmt.Errorf("c.c.PendingNonceAt: %w", err)
	}
	bigAmount := floatToBigInt(amount)

	txValue := bigAmount
	gasLimit := uint64(21000) // in units
	gasPrice, err := c.c.SuggestGasPrice(context.Background())
	if err != nil {
		return "", fmt.Errorf("c.c.SuggestGasPrice: %w", err)
	}

	toAddress := common.HexToAddress(acc)
	var data []byte
	tx := types.NewTransaction(nonce, toAddress, txValue, gasLimit, gasPrice, data)

	chainID, err := c.c.ChainID(context.Background())
	if err != nil {
		return "", fmt.Errorf("c.c.NetworkID: %w", err)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return "", fmt.Errorf("types.SignTx: %w", err)
	}

	err = c.c.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", fmt.Errorf("c.c.SendTransaction; fromAddress: %s: fromAddress; %w", fromAddress, err)
	}

	var bts bytes.Buffer
	bts.WriteString("===\n")
	bts.WriteString("result: success\n")
	bts.WriteString(fmt.Sprintf("tx hash: %s\n", signedTx.Hash().Hex()))
	bts.WriteString("===\n")

	return bts.String(), nil
}

func (c *Client) ChainID() (string, error) {
	chainID, err := c.c.ChainID(context.Background())
	if err != nil {
		return "", err
	}

	return chainID.String(), err
}
