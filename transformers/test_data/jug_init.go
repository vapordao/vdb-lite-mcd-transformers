// VulcanizeDB
// Copyright © 2019 Vulcanize

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

	"github.com/vulcanize/mcd_transformers/transformers/events/jug_init"
	"github.com/vulcanize/mcd_transformers/transformers/shared/constants"
)

var EthJugInitLog = types.Log{
	Address: common.HexToAddress(constants.JugContractAddress()),
	Topics: []common.Hash{
		common.HexToHash(constants.JugInitSignature()),
		common.HexToHash("0x000000000000000000000000dc984d513a0f9ca9aa602d4df8517677918936e3"),
		common.HexToHash("0x434f4c352d410000000000000000000000000000000000000000000000000000"),
		common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"),
	},
	Data:        hexutil.MustDecode("0x000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000e03b663195434f4c352d4100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	BlockNumber: 10980181,
	TxHash:      common.HexToHash("0x27f7834f778ec7d4289cf3337f8e428785c6d023164c02fc44565dbf2e26c49a"),
	TxIndex:     10,
	BlockHash:   fakes.FakeHash,
	Index:       11,
	Removed:     false,
}

var rawJugInitLog, _ = json.Marshal(EthJugInitLog)
var JugInitModel = jug_init.JugInitModel{
	Ilk:              "434f4c352d410000000000000000000000000000000000000000000000000000",
	LogIndex:         EthJugInitLog.Index,
	TransactionIndex: EthJugInitLog.TxIndex,
	Raw:              rawJugInitLog,
}