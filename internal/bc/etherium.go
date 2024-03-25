package bc

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"

	"github.com/vposham/trustdoc/log"
)

// _ maintain this line to force a compilation error when Etherium does not implement OpsIf
var _ OpsIf = (*Etherium)(nil)

// Etherium is an implementation of OpsIf
type Etherium struct {
	httpUrl         string
	from            *common.Address
	privateKey      *ecdsa.PrivateKey
	signer          types.EIP155Signer
	contractAddress *common.Address
	gasLimitOnTx    int64
	gasPrice        *big.Int
	ethCl           *ethclient.Client
	docTkn          *DocumentToken
}

// MintDocTkn creates a new docTkn in Kaleido Etherium private blockchain by using MintDocument method of
// DocumentToken contract
func (k *Etherium) MintDocTkn(ctx context.Context, docId, docHash, ownerEmailHash string) (string, error) {
	logger := log.GetLogger(ctx)
	logger.Info("creating new docTkn", zap.String("docId", docId))
	n, err := k.ethCl.PendingNonceAt(ctx, *k.from)
	if err != nil {
		return "", fmt.Errorf("failed to get nonce: %w", err)

	}
	tx, err := k.docTkn.MintDocument(&bind.TransactOpts{
		From:     *k.contractAddress,
		Nonce:    big.NewInt(int64(n)),
		Signer:   k.sign,
		GasPrice: k.gasPrice,
		GasLimit: uint64(k.gasLimitOnTx),
		Context:  ctx,
	}, docId, docHash, ownerEmailHash)
	if err != nil {
		return "", fmt.Errorf("failed to mint new docTkn: %w", err)
	}

	logger.Info("signed and sent docTkn for mining", zap.Any("bcTxHash", tx.Hash().Hex()))

	// TODO - ideally we shouldnt timelimit on mining and create a event driven system
	// TODO - for these requirements, but for now we will limit the mining time
	// limit time for mining
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	receipt, err := bind.WaitMined(ctx, k.ethCl, tx)
	if err != nil {
		return "", fmt.Errorf("failed to mine docTkn: %w", err)
	}
	logger.Info("new docTkn created, mining complete")

	var docTkn string
	for _, l := range receipt.Logs {
		minted, unpackErr := k.docTkn.ParseDocumentMinted(*l)
		if unpackErr == nil {
			docTkn = minted.TokenId.String()
		}
	}

	if docTkn == "" {
		logger.Error("failed to mint new docTkn", zap.Any("receipt", receipt))
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
	parsedURL, err := url.Parse(k.httpUrl)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
		return
	}

	// Extract username and password from the URL
	username := parsedURL.User.Username()
	password, _ := parsedURL.User.Password()
	hostStr := strings.ReplaceAll(parsedURL.Host, "rpc", "connect")
	getUrl := fmt.Sprintf("%s://%s/gateways/gw/%s/getDocumentContent?_tokenId=%s", parsedURL.Scheme, hostStr,
		k.contractAddress.Hex(), tknId)
	req, err := http.NewRequest("GET", getUrl, nil)
	if err != nil {
		return
	}
	req.SetBasicAuth(username, password)
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.New("docTkn verification failed" + err.Error())
	}
	var target GetDocContentResp
	err = json.NewDecoder(r.Body).Decode(&target)
	if err != nil {
		return err
	}
	if target.Output == docHash {
		logger.Info("docTkn verified")
		return nil
	}
	return errors.New("docTkn verification failed")
}

type GetDocContentResp struct {
	Output string `json:"output"`
}
