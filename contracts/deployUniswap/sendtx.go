package token

import (
	"context"
	"crypto/ecdsa"
	"ethereum/contract/contracts/backends"
	"ethereum/contract/contracts/deployUniswap/cdc"
	factory "ethereum/contract/contracts/deployUniswap/factory"
	"ethereum/contract/contracts/deployUniswap/weth"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	fatallog "log"
	"math"
	"math/big"
	"os"
	"strings"
	"time"
)

var (
	key, _ = crypto.HexToECDSA("956fb3f29e34a14c603f458ffdee4b526a7f6b6b918f6d5f3a9ea7c533fa6b6b")
	addr   = crypto.PubkeyToAddress(key.PublicKey)

	key2, _      = crypto.HexToECDSA("5e6ea3e3ba8a3d8940088247eda01a0909320f729ae3afcdc5747b2ced1ac460")
	testAddr     = crypto.PubkeyToAddress(key2.PublicKey)
	nonce        uint64
	routerAbi, _ = abi.JSON(strings.NewReader(TokenABI))
	cdcAbi, _    = abi.JSON(strings.NewReader(cdc.TokenmABI))
)

func packInput(routerStaking abi.ABI, abiMethod, method string, params ...interface{}) []byte {
	input, err := routerStaking.Pack(abiMethod, params...)
	if err != nil {
		printTest(method, " error ", err)
	}
	return input
}

func printTest(a ...interface{}) {
	log.Info("test", "SendTX", a)
}

func simulateRouter(transactOpts *bind.TransactOpts, contractBackend *backends.SimulatedBackend, basecontract *BaseContract, routercontract *RouterContract) {
	tik := new(big.Int).SetUint64(10000000000000000)
	tik1 := new(big.Int).SetUint64(1000000000000)
	balance, err := basecontract.mapTran.Allowance(nil, addr, routercontract.routerAddr)
	name, err := basecontract.mapTran.Name(nil)
	fmt.Println("balance ", balance, "name", name, " routerAddr ", routercontract.routerAddr.String(), " mapT ", basecontract.mapT.String(), "err", err)
	transactOpts.Value = new(big.Int).SetUint64(1000000000000000000)
	_, err = routercontract.RTran.AddLiquidityETH(transactOpts, basecontract.mapT, tik, tik, tik1, addr, new(big.Int).SetUint64(1699658290))
	fmt.Println("simulate result", err, " routerAddr ", routercontract.rethR.String(), " addr ", addr.String())
	contractBackend.Commit()
}

func sendRouterContract(transactOpts *bind.TransactOpts, contractBackend *backends.SimulatedBackend, client *ethclient.Client, basecontract *BaseContract) (*RouterContract, bool) {
	routerAddr, _, _, err := DeployToken(transactOpts, contractBackend, basecontract.fac, basecontract.weth)
	RTran, err := NewToken(routerAddr, contractBackend)

	_, err = basecontract.mapTran.Approve(transactOpts, routerAddr, new(big.Int).SetUint64(1000000000000000000))
	contractBackend.Commit()
	if err != nil {
		fmt.Println("sendRouterContract", err)
	}
	_, rtx, _, err := DeployToken(transactOpts, contractBackend, basecontract.facR, basecontract.wethR)

	rHash := sendContractTransaction(client, rtx)
	result, rethR := getResult(client, rHash)
	if !result {
		fmt.Println("sendRouterContract getResult", err)
		return nil, false
	}

	input := packInput(cdcAbi, "approve", "approve", rethR, new(big.Int).SetUint64(1000000000000000000))
	aHash := sendRouterTransaction(client, addr, basecontract.mapTR, transactOpts.Value, key, input)
	result, _ = getResult(client, aHash)
	if !result {
		fmt.Println("sendRouterContract getResult", err)
		return nil, false
	}
	return &RouterContract{routerAddr: routerAddr, RTran: RTran, rethR: rethR}, result
}

type RouterContract struct {
	routerAddr common.Address
	RTran      *Token
	rethR      common.Address
}

func sendBaseContract(transactOpts *bind.TransactOpts, contractBackend *backends.SimulatedBackend, client *ethclient.Client) (*BaseContract, bool) {
	weth, wtx, _, err := weth.DeployTokene(transactOpts, contractBackend)
	fac, ftx, _, err := factory.DeployTokenf(transactOpts, contractBackend, addr)
	mapT, mtx, _, err := cdc.DeployTokenc(transactOpts, contractBackend)
	mapTran, err := cdc.NewTokenc(mapT, contractBackend)
	contractBackend.Commit()
	if err != nil {
		fatallog.Fatal("sendBaseContract", err)
	}
	wHash := sendContractTransaction(client, wtx)
	fHash := sendContractTransaction(client, ftx)
	mHash := sendContractTransaction(client, mtx)

	result, wethR := getResult(client, wHash)
	result1, facR := getResult(client, fHash)
	result2, mapTR := getResult(client, mHash)

	if !result || !result1 || !result2 {
		return nil, false
	}
	return &BaseContract{weth: weth, fac: fac, mapT: mapT, mapTran: mapTran,
		wethR: wethR, facR: facR, mapTR: mapTR}, true
}

type BaseContract struct {
	weth, fac, mapT    common.Address
	mapTran            *cdc.Tokenm
	wethR, facR, mapTR common.Address
}

func sendContractTransaction(client *ethclient.Client, txTran *types.Transaction) common.Hash {
	// Ensure a valid value field and resolve the account nonce
	from := addr
	if nonce == 0 {
		var err error
		nonce, err = client.PendingNonceAt(context.Background(), from)
		if err != nil {
			fatallog.Fatal(err)
		}
	} else {
		nonce = nonce + 1
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		fatallog.Fatal(err)
	}

	gasLimit := txTran.Gas()
	value := txTran.Value()

	if gasLimit < 21000 {
		// If the contract surely has code (or code is not needed), estimate the transaction
		msg := ethereum.CallMsg{From: from, To: txTran.To(), GasPrice: gasPrice, Value: value, Data: txTran.Data()}
		gasLimit, err = client.EstimateGas(context.Background(), msg)
		if err != nil {
			fatallog.Fatal("Contract exec failed", err)
		}
	}

	// Create the transaction, sign it and schedule it for execution
	if txTran.To() == nil {

	}
	var tx *types.Transaction
	if txTran.To() == nil {
		tx = types.NewContractCreation(nonce, value, gasLimit, gasPrice, txTran.Data())
	} else {
		tx = types.NewTransaction(nonce, *txTran.To(), value, gasLimit, gasPrice, txTran.Data())
	}

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		fatallog.Fatal(err)
	}
	fmt.Println("TX data nonce ", nonce, " transfer value ", value, " gasLimit ", gasLimit, " gasPrice ", gasPrice, " chainID ", chainID)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), key)
	if err != nil {
		fatallog.Fatal(err)
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		fatallog.Fatal("SendTransaction", err)
	}

	return signedTx.Hash()
}

func sendRouterTransaction(client *ethclient.Client, from, toAddress common.Address, value *big.Int, privateKey *ecdsa.PrivateKey, input []byte) common.Hash {
	// Ensure a valid value field and resolve the account nonce
	nonce = nonce + 1

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		fatallog.Fatal(err)
	}

	gasLimit := uint64(2100000) // in units
	// If the contract surely has code (or code is not needed), estimate the transaction
	msg := ethereum.CallMsg{From: from, To: &toAddress, GasPrice: gasPrice, Value: value, Data: input}
	gasLimit, err = client.EstimateGas(context.Background(), msg)
	if err != nil {
		fatallog.Fatal("Contract exec failed", err)
	}
	if gasLimit < 1 {
		gasLimit = 866328
	}

	// Create the transaction, sign it and schedule it for execution
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, input)

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		fatallog.Fatal(err)
	}
	fmt.Println("TX data nonce ", nonce, " transfer value ", value, " gasLimit ", gasLimit, " gasPrice ", gasPrice, " chainID ", chainID)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		fatallog.Fatal(err)
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		fatallog.Fatal(err)
	}

	return signedTx.Hash()
}
func getAddr( str string )common.Address{
	res:= common.HexToAddress((str))
	return res
}
func getResult(conn *ethclient.Client, txHash common.Hash) (bool, common.Address) {
	fmt.Println("Please waiting ", " txHash ", txHash.String())

	count := 0
	for {
		time.Sleep(time.Millisecond * 400)
		_, isPending, err := conn.TransactionByHash(context.Background(), txHash)
		if err != nil {
			fatallog.Fatal(err)
			return false, common.Address{}
		}
		count++
		if !isPending {
			break
		}
		if count >= 40 {
			fmt.Println("Please use querytx sub command query later.")
			os.Exit(0)
		}
	}

	return queryTx(conn, txHash, false)
}

func queryTx(conn *ethclient.Client, txHash common.Hash, pending bool) (bool, common.Address) {
	if pending {
		_, isPending, err := conn.TransactionByHash(context.Background(), txHash)
		if err != nil {
			fatallog.Fatal(err)
		}
		if isPending {
			println("In tx_pool no validator  process this, please query later")
			os.Exit(0)
		}
	}

	receipt, err := conn.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		fatallog.Fatal(err)
	}

	if receipt.Status == types.ReceiptStatusSuccessful {
		block, err := conn.BlockByHash(context.Background(), receipt.BlockHash)
		if err != nil {
			fatallog.Fatal(err)
		}

		fmt.Println("Transaction Success", " block Number", receipt.BlockNumber.Uint64(), " block txs", len(block.Transactions()), "blockhash", block.Hash().Hex())
		return true, receipt.ContractAddress
	} else if receipt.Status == types.ReceiptStatusFailed {
		fmt.Println("Transaction Failed ", " Block Number", receipt.BlockNumber.Uint64())
	}
	return false, common.Address{}
}

func dialConn(url string) (*ethclient.Client, string) {
	ip := "165.227.99.131"
	port := 8545

	if url == "" {
		url = fmt.Sprintf("http://%s", fmt.Sprintf("%s:%d", ip, port))
	}
	// Create an IPC based RPC connection to a remote node
	// "http://39.100.97.129:8545"
	conn, err := ethclient.Dial(url)
	if err != nil {
		fatallog.Fatal("dialConn", "Failed to connect to the ethereum client: %v", err)
	}
	return conn, url
}

func printBaseInfo(conn *ethclient.Client, url string) *types.Header {
	header, err := conn.HeaderByNumber(context.Background(), nil)
	if err != nil {
		fatallog.Fatal("printBaseInfo", "err", err)
	}

	if common.IsHexAddress(addr.Hex()) {
		fmt.Println("Connect url ", url, " current number ", header.Number.String(), " address ", addr.Hex())
	} else {
		fmt.Println("Connect url ", url, " current number ", header.Number.String())
	}

	return header
}

func PrintBalance(conn *ethclient.Client, from common.Address) {
	balance, err := conn.BalanceAt(context.Background(), from, nil)
	if err != nil {
		log.Error("PrintBalance", "err", err)
	}
	fbalance := new(big.Float)
	fbalance.SetString(balance.String())
	trueValue := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(18)))

	fmt.Println("Your wallet valid balance is ", trueValue, "'eth ")
}

func toEth(val *big.Int) *big.Float {
	baseUnit := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	fbaseUnit := new(big.Float).SetFloat64(float64(baseUnit.Int64()))
	return new(big.Float).Quo(new(big.Float).SetInt(val), fbaseUnit)
}
