/*
 * Copyright 2018 The OpenWallet Authors
 * This file is part of the OpenWallet library.
 *
 * The OpenWallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The OpenWallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package bitcoin

import (
	"errors"
	"fmt"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/imroc/req"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"strings"
)

// Explorer是由bitpay的insight-API提供区块数据查询接口
// 具体接口说明查看https://github.com/bitpay/insight-api
type Explorer struct {
	BaseURL     string
	AccessToken string
	Debug       bool
	client      *req.Req
	//Client *req.Req
}

func NewExplorer(url string, debug bool) *Explorer {
	c := Explorer{
		BaseURL: url,
		//AccessToken: token,
		Debug: debug,
	}

	api := req.New()
	c.client = api

	return &c
}

// Call calls a remote procedure on another node, specified by the path.
func (b *Explorer) Call(path string, request interface{}, method string) (*gjson.Result, error) {

	if b.client == nil {
		return nil, errors.New("API url is not setup. ")
	}

	if b.Debug {
		log.Std.Debug("Start Request API...")
	}

	url := b.BaseURL + path

	r, err := b.client.Do(method, url, request)

	if b.Debug {
		log.Std.Debug("Request API Completed")
	}

	if b.Debug {
		log.Std.Debug("%+v", r)
	}

	if err != nil {
		return nil, err
	}

	resp := gjson.ParseBytes(r.Bytes())
	err = b.isError(&resp)
	if err != nil {
		return nil, err
	}

	//result := resp.Get("result")

	return &resp, nil
}

//isError 是否报错
func (b *Explorer) isError(result *gjson.Result) error {
	var (
		err error
	)

	/*
		//failed 返回错误
		{
			"result": null,
			"error": {
				"code": -8,
				"message": "Block height out of range"
			},
			"id": "foo"
		}
	*/

	if !result.Get("error").Exists() {

		if !result.Exists() {
			return errors.New("Response is empty! ")
		}

		return nil
	}

	errInfo := fmt.Sprintf("[%d]%s",
		result.Get("status").Int(),
		result.Get("error").String())
	err = errors.New(errInfo)

	return err
}

//getBlockByExplorer 获取区块数据
func (wm *WalletManager) getBlockByExplorer(hash string) (*Block, error) {

	path := fmt.Sprintf("block/%s", hash)

	result, err := wm.ExplorerClient.Call(path, nil, "GET")
	if err != nil {
		return nil, err
	}

	return newBlockByExplorer(result), nil
}

//getBlockHashByExplorer 获取区块hash
func (wm *WalletManager) getBlockHashByExplorer(height uint64) (string, error) {

	path := fmt.Sprintf("block-index/%d", height)

	result, err := wm.ExplorerClient.Call(path, nil, "GET")
	if err != nil {
		return "", err
	}

	return result.Get("blockHash").String(), nil
}

//getBlockHeightByExplorer 获取区块链高度
func (wm *WalletManager) getBlockHeightByExplorer() (uint64, error) {

	path := "status"

	result, err := wm.ExplorerClient.Call(path, nil, "GET")
	if err != nil {
		return 0, err
	}

	height := result.Get("info.blocks").Uint()

	return height, nil
}

//getTxIDsInMemPoolByExplorer 获取待处理的交易池中的交易单IDs
func (wm *WalletManager) getTxIDsInMemPoolByExplorer() ([]string, error) {

	return nil, fmt.Errorf("insight-api unsupport query mempool transactions")
}

//GetTransaction 获取交易单
func (wm *WalletManager) getTransactionByExplorer(txid string) (*Transaction, error) {

	path := fmt.Sprintf("tx/%s", txid)

	result, err := wm.ExplorerClient.Call(path, nil, "GET")
	if err != nil {
		return nil, err
	}

	tx := newTxByExplorer(result)

	return tx, nil

}

//listUnspentByExplorer 获取未花交易
func (wm *WalletManager) listUnspentByExplorer(address ...string) ([]*Unspent, error) {

	var (
		utxos = make([]*Unspent, 0)
	)

	addrs := strings.Join(address, ",")

	request := req.Param{
		"addrs": addrs,
	}

	path := "addrs/utxo"

	result, err := wm.ExplorerClient.Call(path, request, "POST")
	if err != nil {
		return nil, err
	}

	array := result.Array()
	for _, a := range array {
		utxos = append(utxos, NewUnspent(&a))
	}

	return utxos, nil

}

func newBlockByExplorer(json *gjson.Result) *Block {

	/*
		{
			"hash": "0000000000002bd2475d1baea1de4067ebb528523a8046d5f9d8ef1cb60460d3",
			"size": 549,
			"height": 1434016,
			"version": 536870912,
			"merkleroot": "ae4310c991ec16cfc7404aaad9fe5fbd533d0b6617c03eb1ac644c89d58b3e18",
			"tx": ["6767a8acc1a63c7978186c582fdea26c47da5e04b0b2b34740a1728bfd959a05", "226dee96373aedd8a3dd00021684b190b7f23f5e16bb186cee11d0560406c19d"],
			"time": 1539066282,
			"nonce": 4089837546,
			"bits": "1a3fffc0",
			"difficulty": 262144,
			"chainwork": "0000000000000000000000000000000000000000000000c6fce84fddeb57e5fb",
			"confirmations": 279,
			"previousblockhash": "0000000000001fdabb5efc93d15ccaf6980642918cd898df6b3ff5fbf26c19c4",
			"nextblockhash": "00000000000024f2bd323157e595613291f83485ddfbbf311323ed0c0dc46545",
			"reward": 0.78125,
			"isMainChain": true,
			"poolInfo": {}
		}
	*/
	obj := &Block{}
	//解析json
	obj.Hash = gjson.Get(json.Raw, "hash").String()
	obj.Confirmations = gjson.Get(json.Raw, "confirmations").Uint()
	obj.Merkleroot = gjson.Get(json.Raw, "merkleroot").String()

	txs := make([]string, 0)
	for _, tx := range gjson.Get(json.Raw, "tx").Array() {
		txs = append(txs, tx.String())
	}

	obj.tx = txs
	obj.Previousblockhash = gjson.Get(json.Raw, "previousblockhash").String()
	obj.Height = gjson.Get(json.Raw, "height").Uint()
	obj.Version = gjson.Get(json.Raw, "version").Uint()
	obj.Time = gjson.Get(json.Raw, "time").Uint()

	return obj
}

func newTxByExplorer(json *gjson.Result) *Transaction {

	/*
			{
			"txid": "9f5eae5b95016825a437ceb9c9224d3e30d3b351f1100e4df5cc0cacac4e668c",
			"version": 1,
			"locktime": 1433760,
			"vin": [],
			"vout": [],
			"blockhash": "0000000000003ac968ee1ae321f35f76d4dcb685045968d60fc39edb20b0eed0",
			"blockheight": 1433761,
			"confirmations": 5,
			"time": 1539050096,
			"blocktime": 1539050096,
			"valueOut": 0.14652549,
			"size": 814,
			"valueIn": 0.14668889,
			"fees": 0.0001634
		}
	*/
	obj := Transaction{}
	//解析json
	obj.TxID = gjson.Get(json.Raw, "txid").String()
	obj.Version = gjson.Get(json.Raw, "version").Uint()
	obj.LockTime = gjson.Get(json.Raw, "locktime").Int()
	obj.BlockHash = gjson.Get(json.Raw, "blockhash").String()
	obj.BlockHeight = gjson.Get(json.Raw, "blockheight").Uint()
	obj.Confirmations = gjson.Get(json.Raw, "confirmations").Uint()
	obj.Blocktime = gjson.Get(json.Raw, "blocktime").Int()
	obj.Size = gjson.Get(json.Raw, "size").Uint()
	obj.Fees = gjson.Get(json.Raw, "fees").String()

	obj.Vins = make([]*Vin, 0)
	if vins := gjson.Get(json.Raw, "vin"); vins.IsArray() {
		for _, vin := range vins.Array() {
			input := newTxVinByExplorer(&vin)
			obj.Vins = append(obj.Vins, input)
		}
	}

	obj.Vouts = make([]*Vout, 0)
	if vouts := gjson.Get(json.Raw, "vout"); vouts.IsArray() {
		for _, vout := range vouts.Array() {
			output := newTxVoutByExplorer(&vout)
			obj.Vouts = append(obj.Vouts, output)
		}
	}

	return &obj
}

func newTxVinByExplorer(json *gjson.Result) *Vin {

	/*
		{
			"txid": "b8c00fff9208cb02f694666084fe0d65c471e92e45cdc3fb2e43af3a772e702d",
			"vout": 0,
			"sequence": 4294967294,
			"n": 0,
			"scriptSig": {
				"hex": "47304402201f77d18435931a6cb51b6dd183decf067f933e92647562f71a33e80988fbc8f6022012abe6824ffa70e5ccb7326e0dbb66144ba71133c1d4a1215da0b17358d7ca660121024d7be1242bd44619779a976cd1cd2d9351fcf58df59929b30a0c69d852302fb5",
				"asm": "304402201f77d18435931a6cb51b6dd183decf067f933e92647562f71a33e80988fbc8f6022012abe6824ffa70e5ccb7326e0dbb66144ba71133c1d4a1215da0b17358d7ca66[ALL] 024d7be1242bd44619779a976cd1cd2d9351fcf58df59929b30a0c69d852302fb5"
			},
			"addr": "msYiUQquCtGucnk3ZaWeJenYmY8WxRoeuv",
			"valueSat": 990000,
			"value": 0.0099,
			"doubleSpentTxID": null
		}
	*/
	obj := Vin{}
	//解析json
	obj.TxID = gjson.Get(json.Raw, "txid").String()
	obj.Vout = gjson.Get(json.Raw, "vout").Uint()
	obj.N = gjson.Get(json.Raw, "n").Uint()
	obj.Addr = gjson.Get(json.Raw, "addr").String()
	obj.Value = gjson.Get(json.Raw, "value").String()
	obj.Coinbase = gjson.Get(json.Raw, "coinbase").String()

	return &obj
}

func newTxVoutByExplorer(json *gjson.Result) *Vout {

	/*
		{
			"value": "0.01652549",
			"n": 0,
			"scriptPubKey": {
				"hex": "76a9142760a760e8d22b5facb380444920e1197f272ea888ac",
				"asm": "OP_DUP OP_HASH160 2760a760e8d22b5facb380444920e1197f272ea8 OP_EQUALVERIFY OP_CHECKSIG",
				"addresses": ["mj7ASAGw8ia2o7Hqvo2XS1d7jGWr5UgEU9"],
				"type": "pubkeyhash"
			},
			"spentTxId": null,
			"spentIndex": null,
			"spentHeight": null
		}
	*/
	obj := Vout{}
	//解析json
	obj.Value = gjson.Get(json.Raw, "value").String()
	obj.N = gjson.Get(json.Raw, "n").Uint()
	obj.ScriptPubKey = gjson.Get(json.Raw, "scriptPubKey,hex").String()

	//提取地址
	if addresses := gjson.Get(json.Raw, "scriptPubKey.addresses"); addresses.IsArray() {
		obj.Addr = addresses.Array()[0].String()
	}

	obj.Type = gjson.Get(json.Raw, "scriptPubKey.type").String()

	return &obj
}

//getBalanceByExplorer 获取地址余额
func (wm *WalletManager) getBalanceByExplorer(address string) (*openwallet.Balance, error) {


	path := fmt.Sprintf("addr/%s?noTxList=1", address)

	result, err := wm.ExplorerClient.Call(path, nil, "GET")
	if err != nil {
		return nil, err
	}

	return newBalanceByExplorer(result), nil
}

func newBalanceByExplorer(json *gjson.Result) *openwallet.Balance {

	/*

	{
		"addrStr": "mnMSQs3HZ5zhJrCEKbqGvcDLjAAxvDJDCd",
		"balance": 3136.82244887,
		"balanceSat": 313682244887,
		"totalReceived": 3136.82244887,
		"totalReceivedSat": 313682244887,
		"totalSent": 0,
		"totalSentSat": 0,
		"unconfirmedBalance": 0,
		"unconfirmedBalanceSat": 0,
		"unconfirmedTxApperances": 0,
		"txApperances": 3909
	}

	*/
	obj := openwallet.Balance{}
	//解析json
	obj.Address = gjson.Get(json.Raw, "addrStr").String()
	obj.Balance = gjson.Get(json.Raw, "balance").String()
	obj.UnconfirmBalance = gjson.Get(json.Raw, "unconfirmedBalance").String()
	u, _ := decimal.NewFromString(obj.UnconfirmBalance)
	b, _ := decimal.NewFromString(obj.UnconfirmBalance)
	obj.ConfirmBalance = b.Sub(u).StringFixed(8)

	return &obj
}
