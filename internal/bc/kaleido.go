package bc

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/vposham/trustdoc/internal/bc/contracts"
	"github.com/vposham/trustdoc/log"
	"go.uber.org/zap"
)

var _ OpsIf = (*Kaleido)(nil)

type Kaleido struct {
	account      common.Address
	privateKey   *ecdsa.PrivateKey
	signer       types.EIP155Signer
	nonce        uint64
	amount       int64
	to           *common.Address
	gasLimitOnTx int64
	gasPrice     *big.Int
	ethCl        *ethclient.Client
	docTkn       *contracts.DocumentToken
}

func (k *Kaleido) MintDocTkn(ctx context.Context, docId, docHash, ownerEmailHash string) (string, error) {
	logger := log.GetLogger(ctx)
	nonce, err := k.ethCl.PendingNonceAt(ctx, k.account)
	if err != nil {
		return "", fmt.Errorf("failed to get nonce: %w", err)
	}
	tx, err := k.docTkn.MintDocument(&bind.TransactOpts{
		From:      common.Address{},
		Nonce:     big.NewInt(int64(nonce)),
		Signer:    k.sign,
		Value:     nil,
		GasPrice:  k.gasPrice,
		GasFeeCap: nil,
		GasTipCap: nil,
		GasLimit:  uint64(k.gasLimitOnTx),
		Context:   ctx,
		NoSend:    false,
	}, docId, docHash, ownerEmailHash)
	if err != nil {
		return "", fmt.Errorf("failed to mint new docTkn: %w", err)
	}
	bcTxHash := tx.Hash().Hex()
	logger.Info("externally signed and sent docTkn for mining", zap.Any("bcTxHash", bcTxHash))
	return bcTxHash, nil
}

func (k *Kaleido) sign(a common.Address, t *types.Transaction) (*types.Transaction, error) {
	return types.SignTx(t, k.signer, k.privateKey)
}
