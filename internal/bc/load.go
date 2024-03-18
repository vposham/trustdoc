// Package bc will contain the interfaces and implementations of blockchain operations
package bc

import (
	"context"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"strconv"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/vposham/trustdoc/internal/bc/contracts"
	"go.uber.org/zap"

	"github.com/vposham/trustdoc/config"
	"github.com/vposham/trustdoc/log"
)

var (
	onceInit      = new(sync.Once)
	concreteImpls = make(map[string]any)
)

const (
	// bcExecKey implementation makes calls to blockchain
	bcExecKey = "bcExecKey"
)

// Load enables us inject this package as dependency from its parent
func Load(ctx context.Context) error {
	var appErr error
	onceInit.Do(func() {
		appErr = loadImpls(ctx)
	})
	return appErr
}

func loadImpls(ctx context.Context) error {
	props := config.GetAll()
	if concreteImpls[bcExecKey] == nil {
		url := props.MustGetString("kaleido.node.api.url")

		// load private signing keys
		privKey := props.MustGetString("kaleido.ext.sign.priv.key")
		signKey, err := ethcrypto.HexToECDSA(privKey)
		if err != nil {
			return fmt.Errorf("failed to parse private key: %w", err)
		}

		// load blockchain transport layer
		rpcClient, err := ethrpc.DialOptions(ctx, url, ethrpc.WithHTTPClient(loadBcHttpClient(ctx)))
		if err != nil {
			return fmt.Errorf("connection to kaliedo blockchain failed: %w", err)
		}

		// get node chainId. This is needed for EIP155 signing
		chainId, err := getNetworkID(ctx, url)
		if err != nil {
			return err
		}

		fromAdd := ethcrypto.PubkeyToAddress(signKey.PublicKey)

		ethCl := ethclient.NewClient(rpcClient)
		nonce, err := ethCl.PendingNonceAt(ctx, fromAdd)
		if err != nil {
			return fmt.Errorf("failed to get nonce: %w", err)
		}

		gasPrice, err := ethCl.SuggestGasPrice(ctx)
		if err != nil {
			return fmt.Errorf("failed to get nonce: %w", err)
		}

		k := Kaleido{
			account:      fromAdd,
			privateKey:   signKey,
			signer:       types.NewEIP155Signer(big.NewInt(chainId)),
			nonce:        nonce,
			amount:       0,
			gasLimitOnTx: props.MustGetInt64("max.gas.per.tx"),
			gasPrice:     gasPrice,
			ethCl:        ethclient.NewClient(rpcClient),
		}

		toKey := props.MustGetString("kaleido.account.key")
		address := common.HexToAddress(toKey)
		k.to = &address

		instance, err := contracts.NewDocumentToken(*k.to, ethCl)
		if err != nil {
			return fmt.Errorf("failed to instantiate a smart contract: %w", err)
		}

		k.docTkn = instance

		var kOps OpsIf = &k
		concreteImpls[bcExecKey] = kOps
	}
	return nil
}

// GetBc is used to get blockchain signing implementation
func GetBc() OpsIf {
	targetImpl := bcExecKey
	v := concreteImpls[targetImpl]
	return v.(OpsIf)
}

// loadBcHttpClient loads all config for http client which talks to blockchain nodes
func loadBcHttpClient(_ context.Context) *http.Client {
	props := config.GetAll()
	return &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:    props.MustGetInt("kaleido.blockchain.http.client.max.conns"),
			MaxConnsPerHost: props.MustGetInt("kaleido.blockchain.http.client.max.conns.per.host"),
			MaxIdleConnsPerHost: props.
				MustGetInt("kaleido.blockchain.http.client.max.idle.conns.per.host"),
			IdleConnTimeout: props.
				MustGetParsedDuration("kaleido.blockchain.http.client.idle.conn.timeout"),
			DialContext: (&net.Dialer{
				KeepAlive: props.MustGetParsedDuration("kaleido.blockchain.http.client.dail.keepalive"),
				Timeout:   props.MustGetParsedDuration("kaleido.blockchain.http.client.dail.timeout"),
			}).DialContext,
			TLSHandshakeTimeout: props.
				MustGetParsedDuration("kaleido.blockchain.http.client.tls.timeout"),
		},
		Timeout: props.MustGetParsedDuration("kaleido.blockchain.http.client.total.timeout"),
	}
}

// getNetworkID returns the network ID from the node
func getNetworkID(ctx context.Context, url string) (int64, error) {
	rpc, err := ethrpc.Dial(url)
	if err != nil {
		return 0, fmt.Errorf("connect to %s failed: %s", url, err)
	}
	defer rpc.Close()
	var strNetworkID string
	err = rpc.Call(&strNetworkID, "net_version")
	if err != nil {
		return 0, fmt.Errorf("failed to query network ID (to use as chain ID in EIP155 signing): %s", err)
	}
	networkID, err := strconv.ParseInt(strNetworkID, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse network ID returned from node '%s': %s", strNetworkID, err)
	}
	log.GetLogger(ctx).Info("get network id", zap.Int64("networkId", networkID))
	return networkID, nil
}
