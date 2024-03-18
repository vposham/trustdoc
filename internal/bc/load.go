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

	"github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
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
		chainId, err := getNetworkID(ctx, url)
		if err != nil {
			return err
		}

		// compile solc smart contract
		mintDocContract, err := compile(ctx)
		if err != nil {
			return err
		}

		k := KaleidoEth{
			CompiledContract:       mintDocContract,
			RPC:                    rpcClient,
			Account:                ethcrypto.PubkeyToAddress(signKey.PublicKey),
			PrivateKey:             signKey,
			Signer:                 types.NewEIP155Signer(big.NewInt(chainId)),
			Nonce:                  0, // we pull the latest nonce
			Amount:                 0,
			ChainId:                chainId,
			RpcTimeout:             props.MustGetParsedDuration("rpc.max.timout.duration"),
			ReceiptWaitMinDuration: props.MustGetParsedDuration("mine.receipt.min.wait.duration"),
			ReceiptWaitMaxDuration: props.MustGetParsedDuration("mine.receipt.max.wait.duration"),
			GasLimitOnTx:           props.MustGetInt64("max.gas.per.tx"),
			GasPrice:               props.MustGetInt64("gas.price"),
		}

		err = k.initializeNonce(k.Account)
		if err != nil {
			return fmt.Errorf("failed to initialize nonce: %w", err)
		}

		// install contract on node.
		// todo - find if there is a way to not install contract on every node on every startup of app
		k.To, err = k.InstallContract()
		if err != nil {
			return fmt.Errorf("failed to install contract: %w", err)
		}

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

func compile(ctx context.Context) (*CompiledSolidity, error) {
	props := config.GetAll()
	method := props.MustGetString("smartcontract.method.name")
	solFilePath := props.MustGetString("smartcontract.solidity.file.path")
	contractName := props.MustGetString("smartcontract.name")
	scErc721NodeModsPath := props.MustGetString("smartcontract.node.modules.path")
	mintDocContract, err := CompileContract(ctx,
		solFilePath, props.MustGetString("smartcontract.evm.version"),
		fmt.Sprintf("%s:%s", solFilePath, contractName), method,
		scErc721NodeModsPath, []string{"string", "string", "string"})
	if err != nil {
		return nil, fmt.Errorf("failed to compile contract: %w", err)
	}
	return mintDocContract, nil
}
