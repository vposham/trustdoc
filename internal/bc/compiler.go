package bc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os/exec"
	"reflect"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/compiler"
	"go.uber.org/zap"

	"github.com/vposham/trustdoc/log"
)

// CompiledSolidity wraps solc compilation of solidity and ABI generation
type CompiledSolidity struct {
	Compiled     string
	ContractInfo compiler.ContractInfo
	PackedCall   []byte
}

// GenerateTypedArgs parses string arguments into a range of types to pass to the ABI call
func GenerateTypedArgs(ctx context.Context, abi abi.ABI, methodName string, strargs []string) ([]interface{}, error) {

	method, exist := abi.Methods[methodName]
	if !exist {
		return nil, fmt.Errorf("method '%s' not found", methodName)
	}

	logger := log.GetLogger(ctx)
	logger.Debug("parsing args for method", zap.String("method", method.String()))
	var typedArgs []interface{}
	for idx, inputArg := range method.Inputs {
		if idx >= len(strargs) {
			return nil, fmt.Errorf("method requires %d args: %s", len(method.Inputs), method)
		}
		strval := strargs[idx]
		switch inputArg.Type.String() {
		case "string":
			typedArgs = append(typedArgs, strval)
		case "int256", "uint256":
			bigInt := big.NewInt(0)
			if _, ok := bigInt.SetString(strval, 10); !ok {
				return nil, fmt.Errorf("could not convert '%s' to %s", strval, inputArg.Type)
			}
			typedArgs = append(typedArgs, bigInt)
		case "bool":
			typedArgs = append(typedArgs, strings.ToLower(strval) == "true")
		case "address":
			if !common.IsHexAddress(strval) {
				return nil, fmt.Errorf("invalid hex address for arg %d: %s", idx, strval)
			}
			typedArgs = append(typedArgs, common.HexToAddress(strval))
		default:
			return nil, fmt.Errorf("no string parsing configured yet for type %s", inputArg.Type)
		}
	}

	return typedArgs, nil

}

type SolcVersion struct {
	Path    string
	Version string
}

var solcVerExtractor = regexp.MustCompile(`\d+\.\d+\.\d+`)

func getSolcVersion(solcPath string) (*SolcVersion, error) {
	cmdOutput := new(bytes.Buffer)
	cmd := exec.Command(solcPath, "--version")
	cmd.Stdout = cmdOutput
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to invoke solc binary '%s' to check version: %s", solcPath, err)
	}
	ver := solcVerExtractor.FindString(cmdOutput.String())
	if ver == "" {
		return nil, fmt.Errorf("failed to extract version from solc '%s' output: %s", solcPath, cmdOutput.String())
	}
	return &SolcVersion{
		Path:    solcPath,
		Version: ver,
	}, nil
}

// CompileContract uses solc to compile the Solidity source and
func CompileContract(ctx context.Context, solidityFile, evmVersion, contractName, method, nodeModulesPath string,
	args []string) (*CompiledSolidity, error) {
	var c CompiledSolidity

	logger := log.GetLogger(ctx)
	solcVer, err := getSolcVersion("solc")
	if err != nil {
		return nil, fmt.Errorf("failed to find solidity version: %s", err)
	}
	solcArgs := []string{
		"--base-path", "./",
		"--combined-json", "bin,bin-runtime,srcmap,srcmap-runtime,abi,userdoc,devdoc,metadata",
		"--optimize",
		"--evm-version", evmVersion,
		"--allow-paths", ".",
		"--include-path", nodeModulesPath,
		solidityFile,
	}
	solOptionsString := strings.Join(append([]string{solcVer.Path}, solcArgs...), " ")
	logger.Info("compiling smart contract using solc", zap.String("solcCmd", solOptionsString))
	cmd := exec.Command(solcVer.Path, solcArgs...)

	// Compile the solidity
	var stderr, stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to compile [%s]: %s", err, stderr.String())
	}

	compiled, err := compiler.ParseCombinedJSON(stdout.Bytes(), "", solcVer.Version, solcVer.Version, solOptionsString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse solc output: %s", err)
	}

	// Check we only have one contract and grab the code/info
	var contract *compiler.Contract
	contractNames := reflect.ValueOf(compiled).MapKeys()

	if contractName != "" {
		if _, ok := compiled[contractName]; !ok {
			return nil, fmt.Errorf("contract %s not found in Solidity file: %s", contractName, contractNames)
		}
		contract = compiled[contractName]
	} else if len(contractNames) != 1 {
		return nil, fmt.Errorf("more than one contract in Solidity file, please set one to call: %s", contractNames)
	} else {
		contractName = contractNames[0].String()
		contract = compiled[contractName]
	}
	c.ContractInfo = contract.Info
	c.Compiled = contract.Code

	// Pack the arguments for calling the contract
	abiJSON, err := json.Marshal(contract.Info.AbiDefinition)
	if err != nil {
		return nil, fmt.Errorf("serializing ABI: %s", err)
	}
	abiJson, err := abi.JSON(bytes.NewReader(abiJSON))
	if err != nil {
		return nil, fmt.Errorf("parsing ABI: %s", err)
	}
	typedArgs, err := GenerateTypedArgs(ctx, abiJson, method, args)
	if err != nil {
		return nil, err
	}
	packedCall, err := abiJson.Pack(method, typedArgs...)
	if err != nil {
		return nil, fmt.Errorf("packing arguments %s for call %s: %s", args, method, err)
	}
	c.PackedCall = packedCall

	return &c, nil
}
