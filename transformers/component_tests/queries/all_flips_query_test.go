package queries

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vulcanize/vulcanizedb/pkg/datastore/postgres"
	"github.com/vulcanize/vulcanizedb/pkg/datastore/postgres/repositories"
	"github.com/vulcanize/vulcanizedb/pkg/fakes"

	"github.com/vulcanize/mcd_transformers/test_config"
	"github.com/vulcanize/mcd_transformers/transformers/component_tests/queries/test_helpers"
	"github.com/vulcanize/mcd_transformers/transformers/events/deal"
	"github.com/vulcanize/mcd_transformers/transformers/events/flip_kick"
	"github.com/vulcanize/mcd_transformers/transformers/test_data"
)

var _ = Describe("All flips view", func() {
	var (
		db              *postgres.DB
		flipKickRepo    flip_kick.FlipKickRepository
		dealRepo        deal.DealRepository
		headerRepo      repositories.HeaderRepository
		contractAddress = "contract address"
		hash1           = common.BytesToHash([]byte{5, 4, 3, 2, 1}).String()
		hash2           = common.BytesToHash([]byte{1, 2, 3, 4, 5}).String()
	)

	BeforeEach(func() {
		db = test_config.NewTestDB(test_config.NewTestNode())
		test_config.CleanTestDB(db)
		flipKickRepo = flip_kick.FlipKickRepository{}
		flipKickRepo.SetDB(db)
		dealRepo = deal.DealRepository{}
		dealRepo.SetDB(db)
		headerRepo = repositories.NewHeaderRepository(db)
		rand.Seed(time.Now().UnixNano())
	})

	AfterEach(func() {
		closeErr := db.Close()
		Expect(closeErr).NotTo(HaveOccurred())
	})

	It("gets the latest state of every bid on the flipper", func() {
		fakeBidId := rand.Int()
		blockOne := rand.Int()
		timestampOne := int(rand.Int31())
		blockTwo := blockOne + 1
		timestampTwo := timestampOne + 1000

		// insert 2 records for the same bid
		blockOneHeader := fakes.GetFakeHeaderWithTimestamp(int64(timestampOne), int64(blockOne))
		blockOneHeader.Hash = hash1
		headerId, headerOneErr := headerRepo.CreateOrUpdateHeader(blockOneHeader)
		Expect(headerOneErr).NotTo(HaveOccurred())
		flipStorageValuesOne := test_helpers.GetFlipStorageValues(1, test_helpers.FakeIlk.Hex, fakeBidId)
		test_helpers.CreateFlip(db, blockOneHeader, flipStorageValuesOne,
			test_helpers.GetFlipMetadatas(strconv.Itoa(fakeBidId)), contractAddress)

		blockTwoHeader := fakes.GetFakeHeaderWithTimestamp(int64(timestampTwo), int64(blockTwo))
		blockTwoHeader.Hash = hash2
		headerTwoId, headerTwoErr := headerRepo.CreateOrUpdateHeader(blockTwoHeader)
		Expect(headerTwoErr).NotTo(HaveOccurred())
		flipStorageValuesTwo := test_helpers.GetFlipStorageValues(2, test_helpers.FakeIlk.Hex, fakeBidId)
		test_helpers.CreateFlip(db, blockTwoHeader, flipStorageValuesTwo,
			test_helpers.GetFlipMetadatas(strconv.Itoa(fakeBidId)), contractAddress)

		ilkId, urnId, setupErr := test_helpers.SetUpFlipBidContext(test_helpers.FlipBidCreationInput{
			DealCreationInput: test_helpers.DealCreationInput{
				Db:              db,
				BidId:           fakeBidId,
				ContractAddress: contractAddress,
			},
			Dealt:            false,
			IlkHex:           test_helpers.FakeIlk.Hex,
			UrnGuy:           test_data.FlipKickModel.Usr,
			FlipKickRepo:     flipKickRepo,
			FlipKickHeaderId: headerId,
		})
		Expect(setupErr).NotTo(HaveOccurred())

		// insert a separate bid with the same ilk
		fakeBidId2 := fakeBidId + 1
		flipStorageValuesThree := test_helpers.GetFlipStorageValues(3, test_helpers.FakeIlk.Hex, fakeBidId2)
		test_helpers.CreateFlip(db, blockTwoHeader, flipStorageValuesThree,
			test_helpers.GetFlipMetadatas(strconv.Itoa(fakeBidId2)), contractAddress)

		expectedBid1 := test_helpers.FlipBidFromValues(strconv.Itoa(fakeBidId), strconv.Itoa(ilkId),
			strconv.Itoa(urnId), "false", blockTwoHeader.Timestamp, blockOneHeader.Timestamp, flipStorageValuesTwo)
		var actualBid1 test_helpers.FlipBid
		queryErr1 := db.Get(&actualBid1, `SELECT bid_id, ilk_id, urn_id, guy, tic, "end", lot, bid, gal, dealt, tab, created, updated FROM api.all_flips($1) WHERE bid_id = $2`,
			test_helpers.FakeIlk.Identifier, fakeBidId)
		Expect(queryErr1).NotTo(HaveOccurred())
		Expect(expectedBid1).To(Equal(actualBid1))

		flipKickErr := test_helpers.CreateFlipKick(contractAddress, fakeBidId2, headerTwoId, flipKickRepo)
		Expect(flipKickErr).NotTo(HaveOccurred())

		expectedBid2 := test_helpers.FlipBidFromValues(strconv.Itoa(fakeBidId2), strconv.Itoa(ilkId),
			strconv.Itoa(urnId), "false", blockTwoHeader.Timestamp, blockTwoHeader.Timestamp, flipStorageValuesThree)
		var actualBid2 test_helpers.FlipBid
		queryErr2 := db.Get(&actualBid2, `SELECT bid_id, ilk_id, urn_id, guy, tic, "end", lot, bid, gal, dealt, tab, created, updated FROM api.all_flips($1) WHERE bid_id = $2`,
			test_helpers.FakeIlk.Identifier, fakeBidId2)
		Expect(queryErr2).NotTo(HaveOccurred())
		Expect(expectedBid2).To(Equal(actualBid2))

		var bidCount int
		countQueryErr := db.Get(&bidCount, `SELECT COUNT(*) FROM api.all_flips($1)`, test_helpers.FakeIlk.Identifier)
		Expect(countQueryErr).NotTo(HaveOccurred())
		Expect(bidCount).To(Equal(2))
	})

	It("ignores bids from other contracts", func() {
		fakeBidId := rand.Int()
		blockNumber := rand.Int()
		timestamp := int(rand.Int31())

		header := fakes.GetFakeHeaderWithTimestamp(int64(timestamp), int64(blockNumber))
		header.Hash = hash1
		headerId, headerOneErr := headerRepo.CreateOrUpdateHeader(header)
		Expect(headerOneErr).NotTo(HaveOccurred())
		flipStorageValues := test_helpers.GetFlipStorageValues(1, test_helpers.FakeIlk.Hex, fakeBidId)
		test_helpers.CreateFlip(
			db, header, flipStorageValues, test_helpers.GetFlipMetadatas(strconv.Itoa(fakeBidId)), contractAddress)

		_, _, setupErr1 := test_helpers.SetUpFlipBidContext(test_helpers.FlipBidCreationInput{
			DealCreationInput: test_helpers.DealCreationInput{
				Db:              db,
				BidId:           fakeBidId,
				ContractAddress: contractAddress,
			},
			Dealt:            false,
			IlkHex:           test_helpers.FakeIlk.Hex,
			UrnGuy:           test_data.FlipKickModel.Usr,
			FlipKickRepo:     flipKickRepo,
			FlipKickHeaderId: headerId,
		})
		Expect(setupErr1).NotTo(HaveOccurred())

		irrelevantBidId := fakeBidId + 1
		irrelevantAddress := "contract address2"
		irrelevantIlkHex := test_helpers.AnotherFakeIlk.Hex
		irrelevantUrn := test_data.FlipKickModel.Gal
		irrelevantFlipValues := test_helpers.GetFlipStorageValues(2, irrelevantIlkHex, irrelevantBidId)
		test_helpers.CreateFlip(db, header, irrelevantFlipValues,
			test_helpers.GetFlipMetadatas(strconv.Itoa(irrelevantBidId)), irrelevantAddress)

		_, _, setupErr2 := test_helpers.SetUpFlipBidContext(test_helpers.FlipBidCreationInput{
			DealCreationInput: test_helpers.DealCreationInput{
				Db:              db,
				BidId:           irrelevantBidId,
				ContractAddress: irrelevantAddress,
			},
			Dealt:            false,
			IlkHex:           irrelevantIlkHex,
			UrnGuy:           irrelevantUrn,
			FlipKickRepo:     flipKickRepo,
			FlipKickHeaderId: headerId,
		})
		Expect(setupErr2).NotTo(HaveOccurred())

		var bidCount int
		countQueryErr := db.Get(&bidCount, `SELECT COUNT(*) FROM api.all_flips($1)`, test_helpers.FakeIlk.Identifier)
		Expect(countQueryErr).NotTo(HaveOccurred())
		Expect(bidCount).To(Equal(1))
	})
})