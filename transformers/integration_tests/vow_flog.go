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

package integration_tests

import (
	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	c2 "github.com/vulcanize/vulcanizedb/libraries/shared/constants"
	fetch "github.com/vulcanize/vulcanizedb/libraries/shared/fetcher"
	"github.com/vulcanize/vulcanizedb/libraries/shared/transformer"
	"github.com/vulcanize/vulcanizedb/pkg/core"
	"github.com/vulcanize/vulcanizedb/pkg/datastore/postgres"

	"github.com/vulcanize/mcd_transformers/test_config"
	"github.com/vulcanize/mcd_transformers/transformers/shared/constants"
	"github.com/vulcanize/mcd_transformers/transformers/test_data"
	"github.com/vulcanize/mcd_transformers/transformers/vow_flog"
	"github.com/vulcanize/vulcanizedb/libraries/shared/factories"
)

var _ = Describe("VowFlog LogNoteTransformer", func() {
	var (
		db         *postgres.DB
		blockChain core.BlockChain
	)

	BeforeEach(func() {
		rpcClient, ethClient, err := getClients(ipc)
		Expect(err).NotTo(HaveOccurred())
		blockChain, err = getBlockChain(rpcClient, ethClient)
		Expect(err).NotTo(HaveOccurred())
		db = test_config.NewTestDB(blockChain.Node())
		test_config.CleanTestDB(db)
	})

	It("transforms VowFlog log events", func() {
		blockNumber := int64(8946819)
		config := transformer.TransformerConfig{
			TransformerName:     constants.VowFlogLabel,
			ContractAddresses:   []string{test_data.KovanVowContractAddress},
			ContractAbi:         test_data.KovanVowABI,
			Topic:               test_data.KovanVowFlogSignature,
			StartingBlockNumber: blockNumber,
			EndingBlockNumber:   blockNumber,
		}

		header, err := persistHeader(db, blockNumber, blockChain)
		Expect(err).NotTo(HaveOccurred())

		fetcher := fetch.NewFetcher(blockChain)
		logs, err := fetcher.FetchLogs(
			transformer.HexStringsToAddresses(config.ContractAddresses),
			[]common.Hash{common.HexToHash(config.Topic)},
			header)
		Expect(err).NotTo(HaveOccurred())

		tr := factories.LogNoteTransformer{
			Config:     config,
			Converter:  &vow_flog.VowFlogConverter{},
			Repository: &vow_flog.VowFlogRepository{},
		}.NewLogNoteTransformer(db)

		err = tr.Execute(logs, header, c2.HeaderMissing)
		Expect(err).NotTo(HaveOccurred())

		var dbResult []vow_flog.VowFlogModel
		err = db.Select(&dbResult, `SELECT era, log_idx, tx_idx from maker.vow_flog`)
		Expect(err).NotTo(HaveOccurred())

		Expect(dbResult[0].Era).To(Equal("1538558052"))
		Expect(dbResult[0].LogIndex).To(Equal(uint(2)))
		Expect(dbResult[0].TransactionIndex).To(Equal(uint(2)))
	})

	It("rechecks vow flog event", func() {
		blockNumber := int64(8946819)
		config := transformer.TransformerConfig{
			TransformerName:     constants.VowFlogLabel,
			ContractAddresses:   []string{test_data.KovanVowContractAddress},
			ContractAbi:         test_data.KovanVowABI,
			Topic:               test_data.KovanVowFlogSignature,
			StartingBlockNumber: blockNumber,
			EndingBlockNumber:   blockNumber,
		}

		header, err := persistHeader(db, blockNumber, blockChain)
		Expect(err).NotTo(HaveOccurred())

		fetcher := fetch.NewFetcher(blockChain)
		logs, err := fetcher.FetchLogs(
			transformer.HexStringsToAddresses(config.ContractAddresses),
			[]common.Hash{common.HexToHash(config.Topic)},
			header)
		Expect(err).NotTo(HaveOccurred())
		tr := factories.LogNoteTransformer{
			Config:     config,
			Converter:  &vow_flog.VowFlogConverter{},
			Repository: &vow_flog.VowFlogRepository{},
		}.NewLogNoteTransformer(db)

		err = tr.Execute(logs, header, c2.HeaderMissing)
		Expect(err).NotTo(HaveOccurred())

		err = tr.Execute(logs, header, c2.HeaderRecheck)
		Expect(err).NotTo(HaveOccurred())

		var headerID int64
		err = db.Get(&headerID, `SELECT id FROM public.headers WHERE block_number = $1`, blockNumber)
		Expect(err).NotTo(HaveOccurred())

		var vowFlogChecked []int
		err = db.Select(&vowFlogChecked, `SELECT vow_flog_checked FROM public.checked_headers WHERE header_id = $1`, headerID)
		Expect(err).NotTo(HaveOccurred())

		Expect(vowFlogChecked[0]).To(Equal(2))

		var dbResult []vow_flog.VowFlogModel
		err = db.Select(&dbResult, `SELECT era, log_idx, tx_idx from maker.vow_flog`)
		Expect(err).NotTo(HaveOccurred())

		Expect(dbResult[0].Era).To(Equal("1538558052"))
		Expect(dbResult[0].LogIndex).To(Equal(uint(2)))
		Expect(dbResult[0].TransactionIndex).To(Equal(uint(2)))
	})
})