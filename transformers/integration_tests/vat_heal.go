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
	"github.com/vulcanize/mcd_transformers/transformers/shared"
	"github.com/vulcanize/vulcanizedb/libraries/shared/constants"
	"github.com/vulcanize/vulcanizedb/libraries/shared/fetcher"
	"github.com/vulcanize/vulcanizedb/libraries/shared/transformer"

	"github.com/vulcanize/mcd_transformers/test_config"
	"github.com/vulcanize/mcd_transformers/transformers/events/vat_heal"
	mcdConstants "github.com/vulcanize/mcd_transformers/transformers/shared/constants"
	"github.com/vulcanize/mcd_transformers/transformers/test_data"
)

var _ = Describe("VatHeal Transformer", func() {
	testVatHealConfig := transformer.EventTransformerConfig{
		TransformerName:   mcdConstants.VatHealLabel,
		ContractAddresses: []string{mcdConstants.VatContractAddress()},
		ContractAbi:       test_data.KovanVatABI,
		Topic:             test_data.KovanVatHealSignature,
	}

	// TODO: Replace block number once there's a heal event on the updated Vat
	XIt("transforms VatHeal log events", func() {
		blockNumber := int64(8935578)
		testVatHealConfig.StartingBlockNumber = blockNumber
		testVatHealConfig.EndingBlockNumber = blockNumber

		rpcClient, ethClient, err := getClients(ipc)
		Expect(err).NotTo(HaveOccurred())
		blockChain, err := getBlockChain(rpcClient, ethClient)
		Expect(err).NotTo(HaveOccurred())

		db := test_config.NewTestDB(blockChain.Node())
		test_config.CleanTestDB(db)

		header, err := persistHeader(db, blockNumber, blockChain)
		Expect(err).NotTo(HaveOccurred())

		logFetcher := fetcher.NewLogFetcher(blockChain)
		logs, err := logFetcher.FetchLogs(
			transformer.HexStringsToAddresses(testVatHealConfig.ContractAddresses),
			[]common.Hash{common.HexToHash(testVatHealConfig.Topic)},
			header)
		Expect(err).NotTo(HaveOccurred())

		tr := shared.LogNoteTransformer{
			Config:     testVatHealConfig,
			Converter:  &vat_heal.VatHealConverter{},
			Repository: &vat_heal.VatHealRepository{},
		}.NewLogNoteTransformer(db)

		err = tr.Execute(logs, header, constants.HeaderMissing)
		Expect(err).NotTo(HaveOccurred())

		var dbResults []vat_heal.VatHealModel
		err = db.Select(&dbResults, `SELECT urn, v, rad from maker.vat_heal`)
		Expect(err).NotTo(HaveOccurred())

		Expect(len(dbResults)).To(Equal(1))
		dbResult := dbResults[0]
		Expect(dbResult.Urn).To(Equal("0000000000000000000000003728e9777b2a0a611ee0f89e00e01044ce4736d1"))
		Expect(dbResult.V).To(Equal("0000000000000000000000003728e9777b2a0a611ee0f89e00e01044ce4736d1"))
		Expect(dbResult.Rad).To(Equal("1000000000000000000000000000"))
	})

	// TODO: Replace block number once there's a heal event on the updated Vat
	XIt("rechecks vat heal event", func() {
		blockNumber := int64(8935578)
		testVatHealConfig.StartingBlockNumber = blockNumber
		testVatHealConfig.EndingBlockNumber = blockNumber

		rpcClient, ethClient, err := getClients(ipc)
		Expect(err).NotTo(HaveOccurred())
		blockChain, err := getBlockChain(rpcClient, ethClient)
		Expect(err).NotTo(HaveOccurred())

		db := test_config.NewTestDB(blockChain.Node())
		test_config.CleanTestDB(db)

		header, err := persistHeader(db, blockNumber, blockChain)
		Expect(err).NotTo(HaveOccurred())

		logFetcher := fetcher.NewLogFetcher(blockChain)
		logs, err := logFetcher.FetchLogs(
			transformer.HexStringsToAddresses(testVatHealConfig.ContractAddresses),
			[]common.Hash{common.HexToHash(testVatHealConfig.Topic)},
			header)
		Expect(err).NotTo(HaveOccurred())

		tr := shared.LogNoteTransformer{
			Config:     testVatHealConfig,
			Converter:  &vat_heal.VatHealConverter{},
			Repository: &vat_heal.VatHealRepository{},
		}.NewLogNoteTransformer(db)

		err = tr.Execute(logs, header, constants.HeaderMissing)
		Expect(err).NotTo(HaveOccurred())

		err = tr.Execute(logs, header, constants.HeaderRecheck)
		Expect(err).NotTo(HaveOccurred())
		var headerID int64
		err = db.Get(&headerID, `SELECT id FROM public.headers WHERE block_number = $1`, blockNumber)
		Expect(err).NotTo(HaveOccurred())

		var vatHealChecked []int
		err = db.Select(&vatHealChecked, `SELECT vat_heal_checked FROM public.checked_headers WHERE header_id = $1`, headerID)
		Expect(err).NotTo(HaveOccurred())

		Expect(vatHealChecked[0]).To(Equal(2))

		var dbResults []vat_heal.VatHealModel
		err = db.Select(&dbResults, `SELECT urn, v, rad from maker.vat_heal`)
		Expect(err).NotTo(HaveOccurred())

		Expect(len(dbResults)).To(Equal(1))
		dbResult := dbResults[0]
		Expect(dbResult.Urn).To(Equal("0000000000000000000000003728e9777b2a0a611ee0f89e00e01044ce4736d1"))
		Expect(dbResult.V).To(Equal("0000000000000000000000003728e9777b2a0a611ee0f89e00e01044ce4736d1"))
		Expect(dbResult.Rad).To(Equal("1000000000000000000000000000"))
	})
})
