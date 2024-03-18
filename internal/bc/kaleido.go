package bc

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
	"go.uber.org/zap"

	"github.com/vposham/trustdoc/internal/bc/contracts"
	"github.com/vposham/trustdoc/log"
)

var _ OpsIf = (*Kaleido)(nil)

type Kaleido struct {
	rpc             *ethrpc.Client
	from            *common.Address
	privateKey      *ecdsa.PrivateKey
	signer          types.EIP155Signer
	amount          int64
	contractAddress *common.Address
	gasLimitOnTx    int64
	gasPrice        *big.Int
	ethCl           *ethclient.Client
	docTkn          *contracts.DocumentToken

	rpcTimeout             time.Duration
	receiptWaitMinDuration time.Duration
	receiptWaitMaxDuration time.Duration
}

func (k *Kaleido) MintDocTkn(ctx context.Context, docId, docHash, ownerEmailHash string) (string, error) {
	logger := log.GetLogger(ctx)
	logger.Info("creating new docTkn", zap.String("docId", docId))
	nonce, err := k.ethCl.PendingNonceAt(ctx, *k.from)
	if err != nil {
		return "", fmt.Errorf("failed contractAddress get nonce: %w", err)
	}
	tx, err := k.docTkn.MintDocument(&bind.TransactOpts{
		From:      *k.contractAddress,
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

func (k *Kaleido) VerifyDocTkn(ctx context.Context, tknId, docMd5Hash, ownerEmailMd5Hash string) (err error) {
	logger := log.GetLogger(ctx)
	logger.Info("verifying a docTkn")
	h := common.HexToHash(tknId)

	bcDocHash, err := k.docTkn.GetDocumentContent(&bind.CallOpts{
		Pending: true,
		From:    *k.contractAddress,
		Context: ctx,
	}, h.Big())
	if err != nil {
		return fmt.Errorf("failed contractAddress verify docTkn: %w", err)
	}

	bcDocOwnerHash, err := k.docTkn.GetDocumentOwner(&bind.CallOpts{
		Pending: true,
		From:    *k.contractAddress,
		Context: ctx,
	}, h.Big())
	if err != nil {
		return fmt.Errorf("failed contractAddress verify docTkn: %w", err)
	}

	if bcDocHash == docMd5Hash && bcDocOwnerHash == ownerEmailMd5Hash {
		logger.Info("docTkn verified")
		return nil
	}
	return errors.New("docTkn verification failed")
}
