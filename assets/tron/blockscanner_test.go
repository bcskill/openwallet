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

package tron

import "testing"

func TestSetRescanBlockHeight(t *testing.T) {
	scanner := NewTronBlockScanner(tw)

	err := scanner.SetRescanBlockHeight(5114310)
	if err != nil {
		t.Errorf("SetRescanBlockHeight failed: %+v", err)
	}
}

func TestScanBlockTask(t *testing.T) {
	scanner := NewTronBlockScanner(tw)
	scanner.ScanBlockTask()
}

func TestGetUnscanRecords(t *testing.T) {
	list, err := tw.GetUnscanRecords()
	if err != nil {
		t.Errorf("GetUnscanRecords failed unexpected error: %v\n", err)
		return
	}

	for _, r := range list {
		t.Logf("GetUnscanRecords unscan: %v", r)
	}
}

func TestTronBlockScanner_RescanFailedRecord(t *testing.T) {
	bs := NewTronBlockScanner(tw)
	bs.RescanFailedRecord()
}

func TestTronBlockScanner_scanning(t *testing.T) {

	//accountID := "WDHupMjR3cR2wm97iDtKajxSPCYEEddoek"
	//address := "miphUAzHHeM1VXGSFnw6owopsQW3jAQZAk"

	//wallet, err := tw.GetWalletInfo(accountID)
	//if err != nil {
	//	t.Errorf("ONTBlockScanner_scanning failed unexpected error: %v\n", err)
	//	return
	//}

	bs := NewTronBlockScanner(tw)

	//bs.DropRechargeRecords(accountID)

	bs.SetRescanBlockHeight(5224300)
	//tw.SaveLocalNewBlock(1355030, "00000000000000125b86abb80b1f94af13a5d9b07340076092eda92dade27686")

	//bs.AddAddress(address, accountID)

	bs.ScanBlockTask()
}
