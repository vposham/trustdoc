// Package bc will contain the interfaces and implementations of blockchain operations
package bc

import (
	"context"
	"errors"
	"fmt"
	"math/big"
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
		wssUrl := props.MustGetString("kaleido.node.wss.api.url")
		logger := log.GetLogger(ctx)

		// load private signing keys
		privKey := props.MustGetString("kaleido.ext.sign.priv.key")
		signKey, err := ethcrypto.HexToECDSA(privKey)
		if err != nil {
			return fmt.Errorf("failed to parse private key: %w", err)
		}

		// load blockchain transport layer
		rpcClient, err := ethrpc.DialOptions(ctx, wssUrl)
		if err != nil {
			return fmt.Errorf("connection to kaliedo blockchain failed: %w", err)
		}

		// get node chainId. This is needed for EIP155 signing
		chainId, err := getNetworkID(ctx, rpcClient)
		if err != nil {
			return err
		}

		fromAdd := ethcrypto.PubkeyToAddress(signKey.PublicKey)
		logger.Info("fromAdd", zap.String("fromAdd", fromAdd.Hex()))

		ethCl := ethclient.NewClient(rpcClient)
		conAddStr := props.GetString("kaleido.smart.contract.address", "")
		// gasPrice, err := ethCl.SuggestGasPrice(ctx)
		// if err != nil {
		// 	return fmt.Errorf("failed to get gas price: %w", err)
		// }

		k := Etherium{
			rpc:                    rpcClient,
			from:                   &fromAdd,
			privateKey:             signKey,
			signer:                 types.NewEIP155Signer(big.NewInt(chainId)),
			contractAddress:        nil,
			gasLimitOnTx:           props.MustGetInt64("max.gas.per.tx"),
			gasPrice:               big.NewInt(10000),
			ethCl:                  ethCl,
			receiptWaitMinDuration: 10 * time.Second,
			receiptWaitMaxDuration: 2 * time.Minute,
			rpcTimeout:             30 * time.Second,
		}

		// var instance *DocumentToken
		if conAddStr == "" {
			auth, err := bind.NewKeyedTransactorWithChainID(signKey, big.NewInt(chainId))
			if err != nil {
				return err
			}
			auth.GasPrice = k.gasPrice
			auth.GasLimit = uint64(k.gasLimitOnTx)
			contractAdd, _, ins, err := DeployDocTknAbi(ctx, auth, ethCl)

			// contractAdd, err := k.InstallContract(ctx)
			if err != nil {
				return fmt.Errorf("failed to deploy new contract: %w", err)
			}
			k.contractAddress = &contractAdd
			k.docTkn = ins
			// conAddStr = contractAdd.Hex()
		}

		// conAdd := common.HexToAddress(conAddStr)

		// if instance == nil {
		// ins, err := NewDocumentToken(*k.contractAddress, ethCl)
		if err != nil {
			return fmt.Errorf("failed to instantiate a smart contract: %w", err)
		}
		// k.docTkn = ins
		// }
		var kOps OpsIf = &k
		concreteImpls[bcExecKey] = kOps
		logger.Info("loaded blockchain implementation")
	}
	return nil
}

// DeployDocTknAbi deploys a new Ethereum contract, binding an instance of Storage to it.
func DeployDocTknAbi(ctx context.Context, auth *bind.TransactOpts, backend *ethclient.Client) (common.Address,
	*types.Transaction, *DocumentToken, error) {
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

	_, err = bind.WaitMined(ctx, backend, tx)
	if err != nil {
		return common.Address{}, nil, nil, fmt.Errorf("failed to install contract: %w", err)
	}

	return address, tx, &DocumentToken{DocumentTokenCaller: DocumentTokenCaller{contract: contract},
		DocumentTokenTransactor: DocumentTokenTransactor{contract: contract},
		DocumentTokenFilterer:   DocumentTokenFilterer{contract: contract}}, nil
}

// GetBc is used to get blockchain ops implementation
func GetBc() OpsIf {
	targetImpl := bcExecKey
	v := concreteImpls[targetImpl]
	return v.(OpsIf)
}

// getNetworkID returns the network ID from the node
func getNetworkID(ctx context.Context, client *ethrpc.Client) (int64, error) {
	var strNetworkID string
	err := client.Call(&strNetworkID, "net_version")
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
