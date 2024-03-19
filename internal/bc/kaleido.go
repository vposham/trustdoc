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
	docTkn          *DocumentToken
	wsUrl           string

	rpcTimeout             time.Duration
	receiptWaitMinDuration time.Duration
	receiptWaitMaxDuration time.Duration

	authTxnOpts *bind.TransactOpts
}

func (k *Kaleido) MintDocTkn(ctx context.Context, docId, docHash, ownerEmailHash string) (string, error) {
	logger := log.GetLogger(ctx)
	logger.Info("creating new docTkn", zap.String("docId", docId))
	nonce, err := k.ethCl.PendingNonceAt(ctx, *k.from)
	if err != nil {
		return "", fmt.Errorf("failed contractAddress get nonce: %w", err)
	}
	tx, err := k.docTkn.MintDocument(&bind.TransactOpts{
		// From:      *k.from,
		Nonce:    big.NewInt(int64(nonce)),
		Signer:   k.sign,
		GasPrice: k.gasPrice,
		GasLimit: uint64(k.gasLimitOnTx),
		Context:  ctx,
		NoSend:   true,
	}, docId, docHash, ownerEmailHash)
	if err != nil {
		return "", fmt.Errorf("failed to mint new docTkn: %w", err)
	}

	bcTxHash, err := k.signAndSendTxn(ctx, tx)
	if err != nil {
		return "", fmt.Errorf("failed to mint new docTkn: %w", err)
	}

	// bcTxHash := tx.Hash().String()
	logger.Info("externally signed and sent docTkn for mining", zap.Any("bcTxHash", bcTxHash))
	// Wait for mining

	receipt, err := k.waitUntilMined(ctx, time.Now(), bcTxHash, 1*time.Second)

	// receipt, err := bind.WaitMined(ctx, k.ethCl, tx)
	if err != nil {
		return "", fmt.Errorf("failed to mine docTkn: %w", err)
	}
	//
	// var event DocumentTokenDocumentMinted
	logger.Info("waiting for DocumentMinted event", zap.Any("receipt", receipt))
	// for _, log := range receipt.Logs {
	// 	err := k.docTkn..(&event, "DocumentMinted", log)
	// 	if err != nil {
	// 		continue // Skip logs that cannot be unpacked as DocumentMinted events
	// 	}

	// Extract the token ID from the event
	// tokenID := event.TokenID
	// fmt.Println("Token ID of the minted document:", tokenID)
	// break // Exit loop after finding the first DocumentMinted event
	// }
	// Set up filter options for the DocumentMinted event
	// watchOpts := &bind.WatchOpts{
	// 	Start:   nil,
	// 	Context: ctx,
	// }

	// Create a channel to receive events
	// events := make(chan *DocumentTokenDocumentMinted, 10)

	// Watch for events using the FilterLogs method
	// sub, err := k.docTkn.WatchDocumentMinted(watchOpts, events, []*big.Int{big.NewInt(0)})
	// if err != nil {
	// 	return "", err
	// }
	// defer sub.Unsubscribe()

	// Process incoming events
	// for {
	// 	select {
	// 	case event := <-events:
	// 		// Log the event data
	// 		fmt.Printf("DocumentMinted event received - Token ID: %v, Doc ID: %v, Doc MD5 Hash: %v\n",
	// 			event.TokenId, event.DocId, event.DocMd5Hash)
	// 		fmt.Println("YAY")
	// 		return event.TokenId.String(), nil
	// 	case err := <-sub.Err():
	// 		return "", err
	// 	}
	// }

	return bcTxHash, nil
}

func (k *Kaleido) sign(a common.Address, t *types.Transaction) (*types.Transaction, error) {
	return types.SignTx(t, k.signer, k.privateKey)
}

func (k *Kaleido) VerifyDocTkn(ctx context.Context, tknId, docMd5Hash, ownerEmailMd5Hash string) (err error) {
	logger := log.GetLogger(ctx)
	logger.Info("verifying a docTkn")
	h := common.HexToHash(tknId)

	fmt.Println("CONTRACT ADD VerifyDocTkn -", k.contractAddress.Hex())

	bcDocHash, err := k.docTkn.GetDocumentContent(&bind.CallOpts{ // all of these need docTknId which is the issue now
		// Pending: true,
		// From:    *k.contractAddress,
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
