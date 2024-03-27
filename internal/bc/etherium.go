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
	"github.com/ethereum/go-ethereum/event"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
	"go.uber.org/zap"

	"github.com/vposham/trustdoc/log"
)

// _ maintain this line to force a compilation error when Etherium does not implement OpsIf
var _ OpsIf = (*Etherium)(nil)

// Etherium is an implementation of OpsIf
type Etherium struct {
	rpc                    *ethrpc.Client
	from                   *common.Address
	privateKey             *ecdsa.PrivateKey
	signer                 types.EIP155Signer
	contractAddress        *common.Address
	gasLimitOnTx           int64
	gasPrice               *big.Int
	ethCl                  *ethclient.Client
	docTkn                 *DocumentToken
	receiptWaitMinDuration time.Duration
	receiptWaitMaxDuration time.Duration
	rpcTimeout             time.Duration
}

// MintDocTkn creates a new docTkn in Kaleido Etherium private blockchain by using MintDocument method of
// DocumentToken contract
func (k *Etherium) MintDocTkn(ctx context.Context, docId, docHash, ownerEmailHash string) (string, error) {
	logger := log.GetLogger(ctx)
	logger.Info("creating new docTkn", zap.String("docId", docId))
	nonce, err := k.ethCl.PendingNonceAt(ctx, *k.from)
	if err != nil {
		return "", fmt.Errorf("failed contractAddress get nonce: %w", err)
	}
	txops := &bind.TransactOpts{
		Nonce:    big.NewInt(int64(nonce)),
		From:     *k.from,
		Signer:   k.sign,
		GasPrice: k.gasPrice,
		GasLimit: uint64(k.gasLimitOnTx),
		Context:  ctx,
	}

	tx, err := k.docTkn.MintDocument(txops, docId, docHash, ownerEmailHash)
	if err != nil {
		return "", fmt.Errorf("failed to mint new docTkn: %w", err)
	}
	ctx, cncl := context.WithTimeout(ctx, 45*time.Second)
	defer cncl()

	// a, _ := DocumentTokenMetaData.GetAbi()

	// xyz := make(chan types.Log)
	// s, err := k.ethCl.SubscribeFilterLogs(ctx, ethereum.FilterQuery{
	// 	BlockHash: nil,
	// 	FromBlock: nil,
	// 	ToBlock:   nil,
	// 	Addresses: []common.Address{*k.contractAddress},
	// 	Topics:    nil,
	// }, xyz)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to watch docTkn: %w", err)
	// }
	// for {
	// 	select {
	// 	case <-ctx.Done():
	// 		return "", fmt.Errorf("context cancelled")
	// 	case <-s.Err():
	// 		return "", fmt.Errorf("watch error")
	// 	case v := <-xyz:
	// 		return v.TxHash.String(), fmt.Errorf("watched")
	// 	}
	// }

	xyz := make(chan *DocumentTokenDocumentMinted)
	var s event.Subscription

	time.Sleep(5 * time.Second)
	s, err = k.docTkn.WatchDocumentMinted(&bind.WatchOpts{Context: ctx}, xyz)
	if err != nil {
		fmt.Println("failed to watch docTkn: %w", err)
	}

	var receipt map[string]interface{}
	err = k.rpc.CallContext(ctx, &receipt, "eth_getTransactionReceipt", tx.Hash().Hex())
	if err != nil {
		return "", fmt.Errorf("failed to mint new docTkn: %w", err)
	}
	logger.Info("signed and sent docTkn for mining", zap.Any("tx", tx), zap.Any("receipt", receipt))

	for {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("context cancelled")
		case <-s.Err():
			return "", fmt.Errorf("watch error")
		case v := <-xyz:
			return v.TokenId.String(), fmt.Errorf("watched")
		}
	}

	// TODO - ideally we shouldnt timelimit on mining and create a event driven system
	// TODO - for these requirements, but for now we will limit the mining time
	// limit time for mining

	// k.docTkn.Sub

	// ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	// defer cancel()
	// fmt.Printf("to %v\n", tx.To())
	// receipt, err := bind.WaitMined(ctx, k.ethCl, tx)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to mine docTkn: %w", err)
	// }
	// logger.Info("new docTkn created, mining complete")

	var docTkn string
	// for _, l := range receipt.Logs {
	// 	minted, unpackErr := k.docTkn.ParseDocumentMinted(*l)
	// 	if unpackErr == nil {
	// 		docTkn = minted.TokenId.String()
	// 	}
	// }

	if docTkn == "" {
		logger.Error("failed to mint new docTkn")
		return docTkn, fmt.Errorf("failed to mint new docTkn")
	}

	return docTkn, nil
}

// sign the transaction
func (k *Etherium) sign(_ common.Address, t *types.Transaction) (*types.Transaction, error) {
	return types.SignTx(t, k.signer, k.privateKey)
}

// VerifyDocTkn verifies a docTkn by comparing the docHash and ownerEmailHash with the one stored in Kaleido Etherium
// private blockchain.
// TODO - can be improved by making only 1 call to blockchain, however it needs a change in the contract
func (k *Etherium) VerifyDocTkn(ctx context.Context, tknId, docHash, ownerEmailHash string) (err error) {
	logger := log.GetLogger(ctx)
	logger.Info("verifying a docTkn", zap.String("docTkn", tknId))
	tkn := new(big.Int)
	tkn, success := tkn.SetString(tknId, 10)
	if !success {
		return fmt.Errorf("failed to parse given tokenId %s", tknId)
	}

	// a, _ := DocumentTokenMetaData.GetAbi() // var out string
	// data, err := a.Pack("getDocumentContent", tkn)
	// if err != nil {
	// 	return fmt.Errorf("failed to pack data: %w", err)
	// }
	// msg := ethereum.CallMsg{From: *k.from, To: k.contractAddress, Data: []byte("0x" + hex.EncodeToString(data))}
	// output, err := k.ethCl.CallContract(ctx, msg, nil)
	// if err != nil {
	// 	return err
	// }
	// resp, err := a.Unpack("getDocumentContent", output)
	// bcDocHash, err := k.docTkn.GetDocumentContent(&bind.CallOpts{
	// 	From:    *k.contractAddress,
	// 	Context: ctx,
	// }, tkn)
	bcDocHash, err := k.docTkn.GetDocumentContent(&bind.CallOpts{
		// Pending: true,
		// From:    *k.contractAddress,
		Context: ctx,
	}, tkn)
	if err != nil {
		return fmt.Errorf("failed contractAddress verify docTkn: %w", err)
	}

	fmt.Println("bcDocHash", bcDocHash)
	// logger.Info("docTkn verified", zap.Any("resp", resp))

	// bcDocOwnerHash, err := k.docTkn.GetDocumentOwner(&bind.CallOpts{
	// 	// Pending: false,
	// 	From:    *k.from,
	// 	Context: ctx,
	// }, tkn)
	// if err != nil {
	// 	return fmt.Errorf("get doc owner - %w", err)
	// }

	// if bcDocHash == docHash {
	// 	logger.Info("docTkn verified")
	// 	return nil
	// }
	return errors.New("docTkn verification failed, contents were altered")
}
