// Package bc will contain the interfaces and implementations of blockchain operations
package bc

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
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
		httpUrl := props.MustGetString("kaleido.node.https.api.url")
		wsUrl := fmt.Sprintf("kaleido.node.wss.api.url")
		logger := log.GetLogger(ctx)

		// load private signing keys
		privKey := props.MustGetString("kaleido.ext.sign.priv.key")
		signKey, err := ethcrypto.HexToECDSA(privKey)
		if err != nil {
			return fmt.Errorf("failed to parse private key: %w", err)
		}

		// load blockchain transport layer
		rpcClient, err := ethrpc.DialOptions(ctx, httpUrl, ethrpc.WithHTTPClient(loadBcHttpClient(ctx)))
		if err != nil {
			return fmt.Errorf("connection to kaliedo blockchain failed: %w", err)
		}

		// get node chainId. This is needed for EIP155 signing
		chainId, err := getNetworkID(ctx, httpUrl)
		if err != nil {
			return err
		}

		fromAdd := ethcrypto.PubkeyToAddress(signKey.PublicKey)
		logger.Info("fromAdd", zap.String("fromAdd", fromAdd.String()))

		ethCl := ethclient.NewClient(rpcClient)

		gasPrice, err := ethCl.SuggestGasPrice(ctx)
		if err != nil {
			return fmt.Errorf("failed to get gas price: %w", err)
		}

		k := Kaleido{
			rpc:                    rpcClient,
			from:                   &fromAdd,
			privateKey:             signKey,
			signer:                 types.NewEIP155Signer(big.NewInt(chainId)),
			amount:                 0,
			contractAddress:        nil, // updated below after contract creation
			gasLimitOnTx:           props.MustGetInt64("max.gas.per.tx"),
			gasPrice:               gasPrice,
			ethCl:                  ethCl,
			docTkn:                 nil,
			rpcTimeout:             45 * time.Second,
			receiptWaitMinDuration: 10 * time.Second,
			receiptWaitMaxDuration: 30 * time.Second,
			wsUrl:                  wsUrl,
		}

		if props.MustGetBool("skip.blockchain.contract.install") {
			var kOps OpsIf = &k
			concreteImpls[bcExecKey] = kOps
			return nil
		}

		// auth, err := bind.NewKeyedTransactorWithChainID(signKey, big.NewInt(chainId))
		// if err != nil {
		// 	return err
		// }
		//
		// contractAdd, tx, instance, err := DeployDocTknAbi(auth, ethCl)
		// if err != nil {
		// 	return fmt.Errorf("failed to deploy new contract: %w", err)
		// }
		//
		// // Wait for the transaction to be mined
		// ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		// defer cancel()
		// conAdd, err := bind.WaitDeployed(ctx, k.ethCl, tx)
		// if err != nil {
		// 	return fmt.Errorf("failed contractAddress wait for mining: %w", err)
		// }
		// logger.Info("contract deployed waited", zap.String("contractAddress", conAdd.String()))
		//
		// k.authTxnOpts = auth
		// k.contractAddress = &contractAdd
		// k.docTkn = instance

		// logger.Info("contractAdd",
		// 	zap.String("contractAdd", contractAdd.String())) // 0x411265504c2267d5fCA3049CB7edB8BE4F377306

		cAdd, err := k.InstallContract(ctx)
		if err != nil {
			return fmt.Errorf("failed contractAddress install contract: %w", err)
		}
		k.contractAddress = cAdd
		logger.Info("contract installed", zap.String("contractAddress", cAdd.String()))

		instance, err := NewDocumentToken(*k.contractAddress, ethCl)
		if err != nil {
			return fmt.Errorf("failed to instantiate a smart contract: %w", err)
		}

		k.docTkn = instance

		var kOps OpsIf = &k
		concreteImpls[bcExecKey] = kOps

		go k.ListenForEvents(ctx)
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

// DeployDocTknAbi deploys a new Ethereum contract, binding an instance of Storage to it.
func DeployDocTknAbi(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction,
	*DocumentToken,
	error) {
	parsed, err := DocumentTokenMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(DocumentTokenMetaData.Bin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	return address, tx, &DocumentToken{DocumentTokenCaller: DocumentTokenCaller{contract: contract},
		DocumentTokenTransactor: DocumentTokenTransactor{contract: contract},
		DocumentTokenFilterer:   DocumentTokenFilterer{contract: contract}}, nil
}
