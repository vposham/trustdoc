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
	"go.uber.org/zap"

	"github.com/vposham/trustdoc/internal/bc/contracts"
	"github.com/vposham/trustdoc/log"
)

// InstallContract installs the contract and returns the address
func (k *Kaleido) InstallContract(ctx context.Context) (*common.Address, error) {
	nonce, err := k.ethCl.PendingNonceAt(context.Background(), *k.from)
	if err != nil {
		return nil, fmt.Errorf("failed contractAddress get nonce: %w", err)
	}
	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: big.NewInt(k.gasPrice.Int64()),
		Gas:      uint64(k.gasLimitOnTx),
		To:       nil, // nil means contract creation
		Value:    big.NewInt(k.amount),
		Data:     common.FromHex(contracts.DocumentTokenMetaData.ABI),
		V:        nil,
		R:        nil,
		S:        nil,
	})
	log.GetLogger(ctx).Info("installing contract...")
	receipt, err := k.sendAndWaitForMining(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed contractAddress install contract: %s", err)
	}
	return receipt.ContractAddress, nil
}

// SendAndWaitForMining sends a single transaction and waits for it contractAddress be mined
func (k *Kaleido) sendAndWaitForMining(ctx context.Context, tx *types.Transaction) (*txnReceipt, error) {
	txHash, err := k.signAndSendTxn(ctx, tx)
	var receipt *txnReceipt
	if err != nil {
		return nil, err
	} else {
		// Wait for mining
		start := time.Now()
		time.Sleep(k.receiptWaitMinDuration)
		receipt, err = k.waitUntilMined(ctx, start, txHash, 1*time.Second)
		if err != nil {
			return nil, fmt.Errorf("failed checking TX receipt: %s", err)
		}
	}
	return receipt, err
}

// signAndSendTx n externally signs and sends a transaction
func (k *Kaleido) signAndSendTxn(ctx context.Context, tx *types.Transaction) (string, error) {
	signedTx, _ := types.SignTx(tx, k.signer, k.privateKey)
	var buff bytes.Buffer
	err := signedTx.EncodeRLP(&buff)
	if err != nil {
		return "", err
	}
	from, _ := types.Sender(k.signer, signedTx)
	if from.Hex() != k.from.Hex() {
		return "", fmt.Errorf("EIP155 signing failed - Account=%s From=%s", k.from.Hex(), from.Hex())
	}

	var txHash string
	data, err := rlp.EncodeToBytes(signedTx)
	if err != nil {
		return txHash, fmt.Errorf("failed contractAddress RLP encode: %s", err)
	}
	err = k.rpc.CallContext(ctx, &txHash, "eth_sendRawTransaction", "0x"+hex.EncodeToString(data))
	return txHash, err
}

// WaitUntilMi ned waits until a given transaction has been mined
func (k *Kaleido) waitUntilMined(ctx context.Context, start time.Time, txHash string,
	retryDelay time.Duration) (*txnReceipt, error) {
	isMined := false
	attempts := 1

	var receipt txnReceipt
	for !isMined {
		err := k.rpcCall(ctx, &receipt, "eth_getTransactionReceipt", common.HexToHash(txHash))
		elapsed := time.Since(start)
		attempts++
		isMined = receipt.BlockNumber != nil && receipt.BlockNumber.ToInt().Uint64() > 0
		if err != nil && !errors.Is(err, ethereum.NotFound) {
			return nil, fmt.Errorf("requesting TX receipt: %s", err)
		}
		if !isMined && elapsed > k.receiptWaitMaxDuration {
			return nil, fmt.Errorf("timed out waiting for tx receipt after %.2fs", elapsed.Seconds())
		}
		if !isMined {
			time.Sleep(retryDelay)
		}
	}
	log.GetLogger(ctx).Info("contract installed and mined", zap.Int("attempts", attempts))
	return &receipt, nil
}

func (k *Kaleido) rpcCall(ctx context.Context, result interface{}, method string, args ...interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, k.rpcTimeout)
	defer cancel()

	err := retry.Do(
		func() error {
			err := k.rpc.CallContext(ctx, result, method, args...)
			return err
		},
		retry.RetryIf(func(err error) bool {
			return strings.Contains(err.Error(), "429")
		}),
		retry.Attempts(10),
		retry.DelayType(func(n uint, err error, config *retry.Config) time.Duration {
			return retry.BackOffDelay(n, err, config)
		}),
	)
	return err
}

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
