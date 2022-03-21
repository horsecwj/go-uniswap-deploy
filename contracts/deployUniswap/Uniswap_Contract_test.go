package token

import (
	"context"
	"ethereum/contract/contracts/backends"
	"ethereum/contract/contracts/deployUniswap/cdc"
	factory "ethereum/contract/contracts/deployUniswap/factory"
	"ethereum/contract/contracts/deployUniswap/weth"
	"fmt"
	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/common/hexutil"
	fatallog "log"

	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/log"
)

func init() {
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlInfo, log.StreamHandler(os.Stderr, log.TerminalFormat(false))))
}

func TestDeployUniswapLocal(t *testing.T) {
	contractBackend := backends.NewSimulatedBackend(core.GenesisAlloc{
		addr:     {Balance: new(big.Int).SetUint64(10000000000000000000)},
		testAddr: {Balance: big.NewInt(100000000000000)}},
		100000000)
	transactOpts := bind.NewKeyedTransactor(key)

	weth, _, _, err := weth.DeployTokene(transactOpts, contractBackend)
	if err != nil {
		t.Fatalf("can't DeployContract: %v", err)
	}
	fac, _, _, err := factory.DeployTokenf(transactOpts, contractBackend, addr)
	if err != nil {
		t.Fatalf("can't DeployContract Factory: %v", err)
	}
	cdcT, _, _, err := cdc.DeployTokenc(transactOpts, contractBackend)
	if err != nil {
		t.Fatalf("can't DeployContract: %v", err)
	}
	cdcTran, err := cdc.NewTokenc(cdcT, contractBackend)
	if err != nil {
		t.Fatalf("can't NewContract: %v", err)
	}

	contractBackend.Commit()

	balance, err := cdcTran.BalanceOf(nil, addr)
	if err != nil {
		log.Error("Failed to retrieve token ", "name: %v", err)
	}
	fmt.Println("addr balance BalanceOf", balance)

	// Deploy the ENS registry
	unisWapAddr, _, _, err := DeployToken(transactOpts, contractBackend, fac, weth)
	if err != nil {
		t.Fatalf("can't DeployContract: %v", err)
	}
	_, err = cdcTran.Approve(transactOpts, unisWapAddr, new(big.Int).SetUint64(1000000000000000000))
	if err != nil {
		t.Fatalf("can't NewContract: %v", err)
	}
	contractBackend.Commit()

	ens, err := NewToken(unisWapAddr, contractBackend)
	if err != nil {
		t.Fatalf("can't NewContract: %v", err)
	}
	tik := new(big.Int).SetUint64(10000000000000000)
	tik1 := new(big.Int).SetUint64(1000000000000)

	balance, err = cdcTran.Allowance(nil, addr, unisWapAddr)
	name, err := cdcTran.Name(nil)
	fmt.Println("balance ", balance, "name", name, " unisWapAddr ", unisWapAddr.String())
	transactOpts.Value = new(big.Int).SetUint64(1000000000000000000)
	fmt.Println(cdcT.String(), " ", addr.String(), " fac ", fac.String(), " eth ", weth.String())
	fmt.Println(hexutil.Encode(tik.Bytes()), " ", hexutil.Encode(tik1.Bytes()))
	backends.SimulateDebug = false
	_, err = ens.AddLiquidityETH(transactOpts, cdcT, tik, tik, tik1, addr, new(big.Int).SetUint64(1699658290))
	if err != nil {
		t.Fatalf("can't NewContract AddLiquidityETH : %v", err)
	}
	contractBackend.Commit()
}

func TestNodeConnect(t *testing.T) {
	url := "http://157.245.118.249:8545"
	client, url := dialConn(url)
	printBaseInfo(client, url)
	PrintBalance(client, addr)

	var err error
	nonce, err = client.PendingNonceAt(context.Background(), addr)
	if err != nil {
		fmt.Println("PendingNonceAt", err)
	}
	fmt.Println("PendingNonceAt", nonce)

	nonce, err = client.NonceAt(context.Background(), addr, nil)
	if err != nil {
		fmt.Println("PendingNonceAt", err)
	}
	fmt.Println(nonce)
}

func TestDeployUniswapWithSimulate(t *testing.T) {
	url := "http://157.245.118.249:8545"

	client, url := dialConn(url)
	printBaseInfo(client, url)
	PrintBalance(client, addr)

	// easy access to transactions get input
	contractBackend := backends.NewSimulatedBackend(core.GenesisAlloc{
		addr:     {Balance: new(big.Int).SetUint64(10000000000000000000)},
		testAddr: {Balance: big.NewInt(100000000000000)}},
		100000000)
	transactOpts := bind.NewKeyedTransactor(key)

	basecontract, result := sendBaseContract(transactOpts, contractBackend, client)
	if !result {
		fmt.Println("sendBaseContract failed")
		return
	}
	routercontract, result := sendRouterContract(transactOpts, contractBackend, client, basecontract)
	if !result {
		fmt.Println("sendBaseContract failed")
		return
	}

	tik := new(big.Int).SetUint64(10000000000000000)
	tik1 := new(big.Int).SetUint64(1000000000000)
	transactOpts.Value = new(big.Int).SetUint64(1000000000000000000)
	input := packInput(routerAbi, "addLiquidityETH", "addLiquidityETH", basecontract.mapTR, tik, tik, tik1, addr, new(big.Int).SetUint64(1699658290))
	aHash := sendRouterTransaction(client, addr, routercontract.rethR, transactOpts.Value, key, input)
	result, _ = getResult(client, aHash)
	fmt.Println("over result", result)
}

func TestDeployUniswapRPC(t *testing.T) {
	url := "https://rinkeby.infura.io/v3/c561706d7070475ab1b59071ee4684b0"

	client, url := dialConn(url)
	printBaseInfo(client, url)
	PrintBalance(client, addr)

	transactOpts := bind.NewKeyedTransactor(key)

	_, wtx, _, err := weth.DeployTokene(transactOpts, client)
	result, wethR := getResult(client, wtx.Hash())

	_, ftx, _, err := factory.DeployTokenf(transactOpts, client, addr)
	_, mtx, _, err := cdc.DeployTokenc(transactOpts, client)
	//weth 合约

	//工厂合约
	result1, facR := getResult(client, ftx.Hash())
	//合约
	result2, mapTR := getResult(client, mtx.Hash())

	if !result || !result1 || !result2 {
		fatallog.Fatal("sendBaseContract", err)
		return
	}

	_, routerTx, _, err := DeployToken(transactOpts, client, facR, wethR)
	result, routerAddr := getResult(client, routerTx.Hash())
	if !result {
		fatallog.Fatal("sendBaseContract routerTx", err)
		return
	}

	mapTran, err := cdc.NewTokenc(mapTR, client)
	atx, err := mapTran.Approve(transactOpts, routerAddr, new(big.Int).SetUint64(1000000000000000000))
	result, _ = getResult(client, atx.Hash())
	if !result {
		fatallog.Fatal("sendBaseContract atx", err)
		return
	}

	tik := new(big.Int).SetUint64(10000000000000000)
	tik1 := new(big.Int).SetUint64(1000000000000)
	transactOpts.Value = new(big.Int).SetUint64(1000000000000000000)
	RTran, err := NewToken(routerAddr, client)
	aHash, _ := RTran.AddLiquidityETH(transactOpts, mapTR, tik, tik, tik1, addr, new(big.Int).SetUint64(1699658290))
	result, _ = getResult(client, aHash.Hash())
	fmt.Println("over result", result)
}
func TestNewTokenCaller(t *testing.T) {
	url := "https://rinkeby.infura.io/v3/c561706d7070475ab1b59071ee4684b0"

	client, url := dialConn(url)
	printBaseInfo(client, url)
	PrintBalance(client, addr)

	transactOpts := bind.NewKeyedTransactor(key)
	_, wtx, _, err := cdc.DeployTokenc(transactOpts, client)

	//_, wtx, _, err := weth.DeployTokene(transactOpts, client)
	result, wethR := getResult(client, wtx.Hash())
	if err != nil{
	}
	if result{
		fmt.Println("wethR", wethR)
	}


}
func  TestDeployToken(t *testing.T) {

	url := "https://rinkeby.infura.io/v3/c561706d7070475ab1b59071ee4684b0"
	client, url := dialConn(url)
	mapTR := getAddr("0x14273ef88AfD5a369db861DC858b137dcA7c5f2C")
	routerAddr := getAddr("0x474BBc7ce78D75FfB424E4C07665F1F3CCb7DD66")
	transactOpts := bind.NewKeyedTransactor(key)

	mapTran, err := cdc.NewTokenc(mapTR, client)


	atx, err := mapTran.Approve(transactOpts, routerAddr, new(big.Int).SetUint64(1000000000000000000))
	result, _ := getResult(client, atx.Hash())
	if !result {
		fatallog.Fatal("sendBaseContract atx", err)
		return
	}

	tik := new(big.Int).SetUint64(10000000000000000)
	tik1 := new(big.Int).SetUint64(1000000000000)
	transactOpts.Value = new(big.Int).SetUint64(1000000000000000000)
	RTran, err := NewToken(routerAddr, client)
	aHash, _ := RTran.AddLiquidityETH(transactOpts, mapTR, tik, tik, tik1, addr, new(big.Int).SetUint64(1699658290))
	result, _ = getResult(client, aHash.Hash())
	fmt.Println("over result", result)
}

func  TestDeployToken2(t *testing.T) {

	url := "https://rinkeby.infura.io/v3/c561706d7070475ab1b59071ee4684b0"
	client, url := dialConn(url)

	printBaseInfo(client, url)
	PrintBalance(client, addr)

	mapTR := getAddr("0x14273ef88AfD5a369db861DC858b137dcA7c5f2C")
	mapTR2 := getAddr("0x4Edd0069b2f3f8d4B8067ABBe2324BbADD56595E")

	routerAddr := getAddr("0x474BBc7ce78D75FfB424E4C07665F1F3CCb7DD66")
	transactOpts := bind.NewKeyedTransactor(key)


	mapTran, err := cdc.NewTokenc(mapTR, client)
	atx, err := mapTran.Approve(transactOpts, routerAddr, new(big.Int).SetUint64(1000000000000000000))
	result, _ := getResult(client, atx.Hash())
	if !result {
		fatallog.Fatal("sendBaseContract atx", err)
		return
	}

	mapTran2, err := cdc.NewTokenc(mapTR2, client)
	atx2, err := mapTran2.Approve(transactOpts, routerAddr, new(big.Int).SetUint64(1000000000000000000))
	result2, _ := getResult(client, atx2.Hash())
	if !result2 {
		fatallog.Fatal("sendBaseContract atx", err)
		return
	}

	tik := new(big.Int).SetUint64(10000000000000000)
	tik1 := new(big.Int).SetUint64(1000000000000)
	transactOpts.Value = new(big.Int).SetUint64(0)

	RTran, err := NewToken(routerAddr, client)
	aHash, _ := RTran.AddLiquidity(transactOpts, mapTR,mapTR2, tik, tik, tik1,tik1, addr, new(big.Int).SetUint64(1699658290))
	result, _ = getResult(client, aHash.Hash())
	//0xd5e7557493c5d7c15f506aebbcafc6764fc5ac5be900c5e501794becad4ae1db
	fmt.Println("over result", result)
}



func  TestDeployToken3(t *testing.T) {

	url := "https://rinkeby.infura.io/v3/c561706d7070475ab1b59071ee4684b0"
	client, url := dialConn(url)

	printBaseInfo(client, url)
	PrintBalance(client, addr)

	mapTR := getAddr("0x14273ef88AfD5a369db861DC858b137dcA7c5f2C")
	mapTR2 := getAddr("0x4Edd0069b2f3f8d4B8067ABBe2324BbADD56595E")

	routerAddr := getAddr("0x474BBc7ce78D75FfB424E4C07665F1F3CCb7DD66")
	transactOpts := bind.NewKeyedTransactor(key)


	mapTran, err := cdc.NewTokenc(mapTR, client)
	atx, err := mapTran.Approve(transactOpts, routerAddr, new(big.Int).SetUint64(1000000000000000000))
	result, _ := getResult(client, atx.Hash())
	if !result {
		fatallog.Fatal("sendBaseContract atx", err)
		return
	}

	//mapTran2, err := cdc.NewTokenc(mapTR2, client)
	//atx2, err := mapTran2.Approve(transactOpts, routerAddr, new(big.Int).SetUint64(1000000000000000000))
	//result2, _ := getResult(client, atx2.Hash())
	//if !result2 {
	//	fatallog.Fatal("sendBaseContract atx", err)
	//	return
	//}

	tik := new(big.Int).SetUint64(10000000000000000)
	tik1 := new(big.Int).SetUint64(1000000000000)
	transactOpts.Value = new(big.Int).SetUint64(0)

	var path []common.Address
	path = append(path, mapTR)
	path = append(path, mapTR2)

	RTran, err := NewToken(routerAddr, client)
	aHash, _ := RTran.SwapExactTokensForTokens(transactOpts, tik, tik1,  path,addr, new(big.Int).SetUint64(1699658290))
	result, _ = getResult(client, aHash.Hash())
	//0x9c2c1b1e0bfff0aec9ff27404617e26a28b39f6afec4a637b2175acebb329398
	fmt.Println("over result", result)
}
// weth factort
//0xe268Cb319c37e3771901069b486aC661Ac68552B
//0xd4Fdd2BEF7785Bf486f373acbA143754d0Fd3e0c

//ERC20
//0x4Edd0069b2f3f8d4B8067ABBe2324BbADD56595E
//0x14273ef88AfD5a369db861DC858b137dcA7c5f2C

//router
//0x474BBc7ce78D75FfB424E4C07665F1F3CCb7DD66

