package bc

import (
	"context"
	"crypto/ecdsa"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
)

var _ OpsIf = (*KaleidoEth)(nil)

type KaleidoEth struct {
	CompiledContract       *CompiledSolidity
	RPC                    *ethrpc.Client
	Account                common.Address
	PrivateKey             *ecdsa.PrivateKey
	Signer                 types.EIP155Signer
	Nonce                  uint64
	Amount                 int64
	ChainId                int64
	To                     *common.Address
	RpcTimeout             time.Duration
	ReceiptWaitMinDuration time.Duration
	ReceiptWaitMaxDuration time.Duration

	GasLimitOnTx int64
	GasPrice     int64
}

func (k KaleidoEth) SignNBurn(ctx context.Context, docId, docHash, ownerEmailHash string) (string, error) {
	tx := k.generateTransaction()
	txHash, err := k.sendTransaction(tx)
	if err != nil {
		return "", err
	}
	return txHash, nil
}
