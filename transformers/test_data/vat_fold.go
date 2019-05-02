// VulcanizeDB
// Copyright © 2018 Vulcanize

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package test_data

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/vulcanize/vulcanizedb/pkg/fakes"

	"github.com/vulcanize/mcd_transformers/transformers/events/vat_fold"
	"github.com/vulcanize/mcd_transformers/transformers/shared/constants"
)

var EthVatFoldLogWithPositiveRate = types.Log{
	Address: common.HexToAddress(constants.VatContractAddress()),
	Topics: []common.Hash{
		common.HexToHash(KovanVatFoldSignature),
		common.HexToHash("0x5245500000000000000000000000000000000000000000000000000000000000"),
		common.HexToHash("0x0000000000000000000000003728e9777b2a0a611ee0f89e00e01044ce4736d1"),
		common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
	},
	Data:        hexutil.MustDecode("0x00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000064e6a6a64d45544800000000000000000000000000000000000000000000000000000000000000000000000000000000003728e9777b2a0a611ee0f89e00e01044ce4736d10000000000000000000000000000000000000000000000000000000000000000"),
	BlockNumber: 8940380,
	TxHash:      common.HexToHash("0xfb37b7a88aa8ad14538d1e244a55939fa07c1828e5ca8168bf4edd56f5fc4d57"),
	TxIndex:     8,
	BlockHash:   fakes.FakeHash,
	Index:       5,
	Removed:     false,
}

var rawVatFoldLogWithPositiveRate, _ = json.Marshal(EthVatFoldLogWithPositiveRate)
var VatFoldModelWithPositiveRate = vat_fold.VatFoldModel{
	Ilk:              "5245500000000000000000000000000000000000000000000000000000000000",
	Urn:              "0x3728e9777B2a0a611ee0F89e00E01044ce4736d1",
	Rate:             "2",
	LogIndex:         EthVatFoldLogWithPositiveRate.Index,
	TransactionIndex: EthVatFoldLogWithPositiveRate.TxIndex,
	Raw:              rawVatFoldLogWithPositiveRate,
}

var EthVatFoldLogWithNegativeRate = types.Log{
	Address: common.HexToAddress(constants.VatContractAddress()),
	Topics: []common.Hash{
		common.HexToHash(KovanVatFoldSignature),
		common.HexToHash("0x66616b6520696c6b000000000000000000000000000000000000000000000000"),
		common.HexToHash("0x0000000000000000000000007d7bee5fcfd8028cf7b00876c5b1421c800561a6"),
		common.HexToHash("0xffffffffffffffffffffffffffffffffffffffffffffffe4e51b291d10b00000"),
	},
	Data:        hexutil.MustDecode("0x000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000e0b65337df66616b6520696c6b0000000000000000000000000000000000000000000000000000000000000000000000007d7bee5fcfd8028cf7b00876c5b1421c800561a6ffffffffffffffffffffffffffffffffffffffffffffffe4e51b291d10b0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	BlockNumber: 75,
	TxHash:      common.HexToHash("0x09fa5041c3046a42947edde6193d70143045c38405842d4b08f6614b09272e76"),
	TxIndex:     0,
	BlockHash:   common.HexToHash("0x843138ef186be9695fbd9bbde858491a7d324735175f3b3d4d8e228fa8423271"),
	Index:       0,
	Removed:     false,
}

var rawVatFoldLogWithNegativeRate, _ = json.Marshal(EthVatFoldLogWithNegativeRate)
var VatFoldModelWithNegativeRate = vat_fold.VatFoldModel{
	Ilk:              "66616b6520696c6b000000000000000000000000000000000000000000000000",
	Urn:              "0x7d7bEe5fCfD8028cf7b00876C5b1421c800561A6",
	Rate:             "-500000000000000000000",
	LogIndex:         0,
	TransactionIndex: 0,
	Raw:              rawVatFoldLogWithNegativeRate,
}
