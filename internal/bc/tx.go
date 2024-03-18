package bc

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	retry "github.com/avast/retry-go"
	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

type txnReceipt struct {
	BlockHash         *common.Hash    `json:"blockHash"`
	BlockNumber       *hexutil.Big    `json:"blockNumber"`
	ContractAddress   *common.Address `json:"contractAddress"`
	CumulativeGasUsed *hexutil.Big    `json:"cumulativeGasUsed"`
	TransactionHash   *common.Hash    `json:"transactionHash"`
	From              *common.Address `json:"from"`
	GasUsed           *hexutil.Big    `json:"gasUsed"`
	Status            *hexutil.Big    `json:"status"`
	To                *common.Address `json:"to"`
	TransactionIndex  *hexutil.Uint   `json:"transactionIndex"`
}

// SendAndWaitForMining sends a single transaction and waits for it to be mined
func (k *KaleidoEth) sendAndWaitForMining(tx *types.Transaction) (*txnReceipt, error) {
	txHash, err := k.sendTransaction(tx)
	var receipt *txnReceipt
	if err != nil {
		k.error("failed sending TX: %s", err)
	} else {
		// Wait for mining
		start := time.Now()
		k.debug("Waiting for %d seconds for tx be mined in next block", k.ReceiptWaitMinDuration)
		time.Sleep(k.ReceiptWaitMinDuration)
		receipt, err = k.waitUntilMined(start, txHash, 1*time.Second)
		if err != nil {
			return nil, fmt.Errorf("failed checking TX receipt: %s", err)
		}
		// Increase nonce only if we got a receipt.
		// Known transaction processing will kick in to bump the nonce otherwise
		k.Nonce++
	}
	return receipt, err
}

// sendTransaction sends an individual transaction, choosing external or internal signing
func (k *KaleidoEth) sendTransaction(tx *types.Transaction) (string, error) {
	start := time.Now()

	var err error
	var txHash string
	txHash, err = k.signAndSendTxn(tx)

	callTime := time.Since(start)
	ok := err == nil

	if !ok && (strings.Contains(err.Error(), "known transaction") || strings.Contains(err.Error(), "nonce too low")) {
		// Bump the nonce for the next attempt
		k.Nonce++
	}

	k.info("TX:%s Sent. OK=%t [%.2fs]", txHash, ok, callTime.Seconds())
	return txHash, err
}

// signAndSendTx n externally signs and sends a transaction
func (k *KaleidoEth) signAndSendTxn(tx *types.Transaction) (string, error) {
	signedTx, _ := types.SignTx(tx, k.Signer, k.PrivateKey)
	var buff bytes.Buffer
	err := signedTx.EncodeRLP(&buff)
	if err != nil {
		return "", err
	}
	from, _ := types.Sender(k.Signer, signedTx)
	k.debug("TX signed. ChainID=%d From=%s", k.ChainId, from.Hex())
	if from.Hex() != k.Account.Hex() {
		return "", fmt.Errorf("EIP155 signing failed - Account=%s From=%s", k.Account.Hex(), from.Hex())
	}

	var txHash string
	data, err := rlp.EncodeToBytes(signedTx)
	if err != nil {
		return txHash, fmt.Errorf("failed to RLP encode: %s", err)
	}
	err = k.rpcCall(&txHash, ethRawTx, "0x"+hex.EncodeToString(data))
	return txHash, err
}

func (k *KaleidoEth) rpcCall(result interface{}, method string, args ...interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), k.RpcTimeout)
	defer cancel()

	err := retry.Do(
		func() error {
			k.debug("Invoking %s", method)
			err := k.RPC.CallContext(ctx, result, method, args...)
			return err
		},
		retry.RetryIf(func(err error) bool {
			return strings.Contains(err.Error(), "429")
		}),
		retry.Attempts(10),
		retry.DelayType(func(n uint, err error, config *retry.Config) time.Duration {
			k.debug("%s attempt %d failed: %s", method, n, err)
			return retry.BackOffDelay(n, err, config)
		}),
	)
	return err
}

// initializeN once get the initial nonce to use
func (k *KaleidoEth) initializeNonce(address common.Address) error {
	block := "latest"
	if k.Nonce == 0 {
		var result hexutil.Uint64
		err := k.RPC.Call(&result, ethGetTxCount, address.Hex(), block)
		if err != nil {
			return fmt.Errorf("failed to get transaction count '%s' for %s: %s", result, address, err)
		}
		k.debug("Received nonce=%d for %s at '%s' block", result, address.Hex(), block)
		k.Nonce = uint64(result)
	}
	return nil
}

// WaitUntilMi ned waits until a given transaction has been mined
func (k *KaleidoEth) waitUntilMined(start time.Time, txHash string, retryDelay time.Duration) (*txnReceipt, error) {

	var isMined = false

	var receipt txnReceipt
	for !isMined {
		callStart := time.Now()

		err := k.rpcCall(&receipt, ethGetTxReceipt, common.HexToHash(txHash))
		elapsed := time.Since(start)
		callTime := time.Since(callStart)

		isMined = receipt.BlockNumber != nil && receipt.BlockNumber.ToInt().Uint64() > 0
		k.info("TX:%s Mined=%t after %.2fs [%.2fs]", txHash, isMined, elapsed.Seconds(), callTime.Seconds())
		if err != nil && !errors.Is(err, ethereum.NotFound) {
			return nil, fmt.Errorf("requesting TX receipt: %s", err)
		}
		if receipt.Status != nil {
			status := receipt.Status.ToInt()
			k.debug("Status=%s BlockNumber=%s BlockHash=%x TransactionIndex=%d GasUsed=%s CumulativeGasUsed=%s",
				status, receipt.BlockNumber.ToInt(), receipt.BlockHash,
				receipt.TransactionIndex, receipt.GasUsed.ToInt(), receipt.CumulativeGasUsed.ToInt())
		}
		if !isMined && elapsed > k.ReceiptWaitMaxDuration {
			return nil, fmt.Errorf("timed out waiting for TX receipt after %.2fs", elapsed.Seconds())
		}
		if !isMined {
			time.Sleep(retryDelay)
		}
	}
	return &receipt, nil
}

// InstallContract installs the contract and returns the address
func (k *KaleidoEth) InstallContract() (*common.Address, error) {
	tx := types.NewTx(&types.LegacyTx{
		Nonce:    k.Nonce,
		GasPrice: big.NewInt(k.GasPrice),
		Gas:      uint64(k.GasLimitOnTx),
		To:       nil, // nil means contract creation
		Value:    big.NewInt(k.Amount),
		Data:     common.FromHex(k.CompiledContract.Compiled),
		V:        nil,
		R:        nil,
		S:        nil,
	})
	receipt, err := k.sendAndWaitForMining(tx)
	if err != nil {
		return nil, fmt.Errorf("failed to install contract: %s", err)
	}
	return receipt.ContractAddress, nil
}

// generateTransaction creates a new transaction for the specified data
func (k *KaleidoEth) generateTransaction() *types.Transaction {
	tx := types.NewTx(&types.LegacyTx{
		Nonce:    k.Nonce,
		GasPrice: big.NewInt(k.GasPrice),
		Gas:      uint64(k.GasLimitOnTx),
		To:       k.To,
		Value:    big.NewInt(k.Amount),
		Data:     common.FromHex(k.CompiledContract.Compiled),
		V:        nil,
		R:        nil,
		S:        nil,
	})
	k.debug("TX:%s To=%s Amount=%d Gas=%d GasPrice=%d",
		tx.Hash().Hex(), tx.To().Hex(), k.Amount, k.GasLimitOnTx, k.GasPrice)
	return tx
}

func (k KaleidoEth) debug(message string, inserts ...interface{}) {
	fmt.Printf("%06d:  %s\n", k.Nonce, fmt.Sprintf(message, inserts...))
}

func (k KaleidoEth) info(message string, inserts ...interface{}) {
	fmt.Printf("%06d:  %s\n", k.Nonce, fmt.Sprintf(message, inserts...))
}

func (k KaleidoEth) error(message string, inserts ...interface{}) {
	fmt.Printf("%06d: %s\n", k.Nonce, fmt.Sprintf(message, inserts...))
}
