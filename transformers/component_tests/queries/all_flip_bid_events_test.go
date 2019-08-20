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

package queries

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vulcanize/mcd_transformers/test_config"
	"github.com/vulcanize/mcd_transformers/transformers/component_tests/queries/test_helpers"
	"github.com/vulcanize/mcd_transformers/transformers/events/deal"
	"github.com/vulcanize/mcd_transformers/transformers/events/dent"
	"github.com/vulcanize/mcd_transformers/transformers/events/flip_kick"
	"github.com/vulcanize/mcd_transformers/transformers/events/flip_tick"
	"github.com/vulcanize/mcd_transformers/transformers/events/tend"
	"github.com/vulcanize/mcd_transformers/transformers/events/yank"
	"github.com/vulcanize/mcd_transformers/transformers/storage"
	"github.com/vulcanize/mcd_transformers/transformers/test_data"
	"github.com/vulcanize/vulcanizedb/pkg/datastore/postgres"
	"github.com/vulcanize/vulcanizedb/pkg/datastore/postgres/repositories"
	"github.com/vulcanize/vulcanizedb/pkg/fakes"
	"math/rand"
	"strconv"
)

var _ = Describe("All flip bid events query", func() {
	var (
		db              *postgres.DB
		flipKickRepo    flip_kick.FlipKickRepository
		tendRepo        tend.TendRepository
		flipTickRepo    flip_tick.FlipTickRepository
		dentRepo        dent.DentRepository
		dealRepo        deal.DealRepository
		yankRepo        yank.YankRepository
		headerRepo      repositories.HeaderRepository
		contractAddress = "FlipContract"
		bidId           int
	)

	BeforeEach(func() {
		db = test_config.NewTestDB(test_config.NewTestNode())
		test_config.CleanTestDB(db)
		headerRepo = repositories.NewHeaderRepository(db)
		flipKickRepo = flip_kick.FlipKickRepository{}
		flipKickRepo.SetDB(db)
		tendRepo = tend.TendRepository{}
		tendRepo.SetDB(db)
		flipTickRepo = flip_tick.FlipTickRepository{}
		flipTickRepo.SetDB(db)
		dentRepo = dent.DentRepository{}
		dentRepo.SetDB(db)
		dealRepo = deal.DealRepository{}
		dealRepo.SetDB(db)
		yankRepo = yank.YankRepository{}
		yankRepo.SetDB(db)
		bidId = rand.Int()
	})

	AfterEach(func() {
		closeErr := db.Close()
		Expect(closeErr).NotTo(HaveOccurred())
	})

	Describe("all_flip_bid_events", func() {
		It("returns all flip bid events when they are all in the same block", func() {
			tendLot := rand.Int()
			tendBidAmount := rand.Int()
			dentLot := rand.Int()
			dentBidAmount := rand.Int()

			headerOne := fakes.GetFakeHeader(1)
			headerOneId, headerOneErr := headerRepo.CreateOrUpdateHeader(headerOne)
			Expect(headerOneErr).NotTo(HaveOccurred())

			flipKickEvent := test_data.FlipKickModel
			flipKickEvent.ContractAddress = contractAddress
			flipKickEvent.BidId = strconv.Itoa(bidId)
			flipKickErr := flipKickRepo.Create(headerOneId, []interface{}{flipKickEvent})
			Expect(flipKickErr).NotTo(HaveOccurred())

			flipTendErr := test_helpers.CreateTend(test_helpers.TendCreationInput{
				BidId:           bidId,
				ContractAddress: contractAddress,
				Lot:             tendLot,
				BidAmount:       tendBidAmount,
				TendRepo:        tendRepo,
				TendHeaderId:    headerOneId,
			})
			Expect(flipTendErr).NotTo(HaveOccurred())

			flipTickErr := test_helpers.CreateFlipTick(test_helpers.FlipTickCreationInput{
				BidId:            bidId,
				ContractAddress:  contractAddress,
				FlipTickRepo:     flipTickRepo,
				FlipTickHeaderId: headerOneId,
			})
			Expect(flipTickErr).NotTo(HaveOccurred())

			flipStorageValues := test_helpers.GetFlipStorageValues(1, test_helpers.FakeIlk.Hex, bidId)
			test_helpers.CreateFlip(db, headerOne, flipStorageValues,
				test_helpers.GetFlipMetadatas(strconv.Itoa(bidId)), contractAddress)

			flipDentErr := test_helpers.CreateDent(test_helpers.DentCreationInput{
				BidId:           bidId,
				ContractAddress: contractAddress,
				Lot:             dentLot,
				BidAmount:       dentBidAmount,
				DentRepo:        dentRepo,
				DentHeaderId:    headerOneId,
			})
			Expect(flipDentErr).NotTo(HaveOccurred())

			var actualBidEvents []test_helpers.BidEvent
			queryErr := db.Select(&actualBidEvents, `SELECT bid_id, bid_amount, lot, act FROM api.all_flip_bid_events()`)
			Expect(queryErr).NotTo(HaveOccurred())

			Expect(actualBidEvents).To(ConsistOf(
				test_helpers.BidEvent{BidId: flipKickEvent.BidId, BidAmount: flipKickEvent.Bid, Lot: flipKickEvent.Lot, Act: "kick"},
				test_helpers.BidEvent{BidId: strconv.Itoa(bidId), BidAmount: strconv.Itoa(tendBidAmount), Lot: strconv.Itoa(tendLot), Act: "tend"},
				test_helpers.BidEvent{BidId: strconv.Itoa(bidId), BidAmount: flipStorageValues[storage.BidBid].(string), Lot: flipStorageValues[storage.BidLot].(string), Act: "tick"},
				test_helpers.BidEvent{BidId: strconv.Itoa(bidId), BidAmount: strconv.Itoa(dentBidAmount), Lot: strconv.Itoa(dentLot), Act: "dent"}))
		})

		It("returns flip bid events across all blocks", func() {
			tendLot := rand.Int()
			tendBidAmount := rand.Int()
			dentLot := rand.Int()
			dentBidAmount := rand.Int()

			headerOne := fakes.GetFakeHeader(1)
			headerOneId, headerOneErr := headerRepo.CreateOrUpdateHeader(headerOne)
			Expect(headerOneErr).NotTo(HaveOccurred())

			flipKickEvent := test_data.FlipKickModel
			flipKickEvent.ContractAddress = contractAddress
			flipKickEvent.BidId = strconv.Itoa(bidId)
			flipKickErr := flipKickRepo.Create(headerOneId, []interface{}{flipKickEvent})
			Expect(flipKickErr).NotTo(HaveOccurred())

			flipTendErr := test_helpers.CreateTend(test_helpers.TendCreationInput{
				BidId:           bidId,
				ContractAddress: contractAddress,
				Lot:             tendLot,
				BidAmount:       tendBidAmount,
				TendRepo:        tendRepo,
				TendHeaderId:    headerOneId,
			})
			Expect(flipTendErr).NotTo(HaveOccurred())

			headerTwo := fakes.GetFakeHeader(2)
			headerTwoId, headerTwoErr := headerRepo.CreateOrUpdateHeader(headerTwo)
			Expect(headerTwoErr).NotTo(HaveOccurred())

			flipTickErr := test_helpers.CreateFlipTick(test_helpers.FlipTickCreationInput{
				BidId:            bidId,
				ContractAddress:  contractAddress,
				FlipTickRepo:     flipTickRepo,
				FlipTickHeaderId: headerTwoId,
			})
			Expect(flipTickErr).NotTo(HaveOccurred())

			flipStorageValuesBlockTwo := test_helpers.GetFlipStorageValues(2, test_helpers.FakeIlk.Hex, bidId)
			test_helpers.CreateFlip(db, headerTwo, flipStorageValuesBlockTwo,
				test_helpers.GetFlipMetadatas(strconv.Itoa(bidId)), contractAddress)

			headerThree := fakes.GetFakeHeader(3)
			headerThreeId, headerThreeErr := headerRepo.CreateOrUpdateHeader(headerThree)
			Expect(headerThreeErr).NotTo(HaveOccurred())

			flipDentErr := test_helpers.CreateDent(test_helpers.DentCreationInput{
				BidId:           bidId,
				ContractAddress: contractAddress,
				Lot:             dentLot,
				BidAmount:       dentBidAmount,
				DentRepo:        dentRepo,
				DentHeaderId:    headerThreeId,
			})
			Expect(flipDentErr).NotTo(HaveOccurred())

			flipDealErr := test_helpers.CreateDeal(test_helpers.DealCreationInput{
				Db:              db,
				BidId:           bidId,
				ContractAddress: contractAddress,
				DealRepo:        dealRepo,
				DealHeaderId:    headerThreeId,
			})
			Expect(flipDealErr).NotTo(HaveOccurred())

			flipStorageValuesBlockThree := test_helpers.GetFlipStorageValues(3, test_helpers.FakeIlk.Hex, bidId)
			test_helpers.CreateFlip(db, headerThree, flipStorageValuesBlockThree,
				test_helpers.GetFlipMetadatas(strconv.Itoa(bidId)), contractAddress)

			var actualBidEvents []test_helpers.BidEvent
			queryErr := db.Select(&actualBidEvents, `SELECT bid_id, bid_amount, lot, act FROM api.all_flip_bid_events()`)
			Expect(queryErr).NotTo(HaveOccurred())

			Expect(actualBidEvents).To(ConsistOf(
				test_helpers.BidEvent{BidId: flipKickEvent.BidId, BidAmount: flipKickEvent.Bid, Lot: flipKickEvent.Lot, Act: "kick"},
				test_helpers.BidEvent{BidId: strconv.Itoa(bidId), BidAmount: strconv.Itoa(tendBidAmount), Lot: strconv.Itoa(tendLot), Act: "tend"},
				test_helpers.BidEvent{BidId: strconv.Itoa(bidId), BidAmount: flipStorageValuesBlockTwo[storage.BidBid].(string), Lot: flipStorageValuesBlockTwo[storage.BidLot].(string), Act: "tick"},
				test_helpers.BidEvent{BidId: strconv.Itoa(bidId), BidAmount: strconv.Itoa(dentBidAmount), Lot: strconv.Itoa(dentLot), Act: "dent"},
				test_helpers.BidEvent{BidId: strconv.Itoa(bidId), BidAmount: flipStorageValuesBlockThree[storage.BidBid].(string), Lot: flipStorageValuesBlockThree[storage.BidLot].(string), Act: "deal"}))
		})

		It("returns bid events from flippers that have different bid ids", func() {
			differentBidId := rand.Int()
			differentLot := rand.Int()

			headerOne := fakes.GetFakeHeader(1)
			headerOneId, headerOneErr := headerRepo.CreateOrUpdateHeader(headerOne)
			Expect(headerOneErr).NotTo(HaveOccurred())

			flipKickEventOne := test_data.FlipKickModel
			flipKickEventOne.ContractAddress = contractAddress
			flipKickEventOne.BidId = strconv.Itoa(bidId)
			flipKickErr := flipKickRepo.Create(headerOneId, []interface{}{flipKickEventOne})
			Expect(flipKickErr).NotTo(HaveOccurred())

			flipKickEventTwo := test_data.FlipKickModel
			flipKickEventTwo.ContractAddress = contractAddress
			flipKickEventTwo.BidId = strconv.Itoa(differentBidId)
			flipKickEventTwo.Lot = strconv.Itoa(differentLot)
			flipKickEventTwo.TransactionIndex = flipKickEventOne.TransactionIndex + 1
			flipKickEventTwo.LogIndex = flipKickEventOne.LogIndex + 1
			flipKickErr = flipKickRepo.Create(headerOneId, []interface{}{flipKickEventTwo})
			Expect(flipKickErr).NotTo(HaveOccurred())

			var actualBidEvents []test_helpers.BidEvent
			queryErr := db.Select(&actualBidEvents, `SELECT bid_id, bid_amount, lot, act FROM api.all_flip_bid_events()`)
			Expect(queryErr).NotTo(HaveOccurred())

			Expect(actualBidEvents).To(ConsistOf(
				test_helpers.BidEvent{BidId: flipKickEventOne.BidId, BidAmount: flipKickEventOne.Bid, Lot: flipKickEventOne.Lot, Act: "kick"},
				test_helpers.BidEvent{BidId: flipKickEventTwo.BidId, BidAmount: flipKickEventTwo.Bid, Lot: flipKickEventTwo.Lot, Act: "kick"},
			))
		})

		It("returns bid events from different kinds of flips (flips with different contract addresses", func() {
			anotherFlipContractAddress := "different flip contract address"
			differentLot := rand.Int()
			differentBidAmount := rand.Int()

			headerOne := fakes.GetFakeHeader(1)
			headerOneId, headerOneErr := headerRepo.CreateOrUpdateHeader(headerOne)
			Expect(headerOneErr).NotTo(HaveOccurred())

			flipKickEventOne := test_data.FlipKickModel
			flipKickEventOne.ContractAddress = contractAddress
			flipKickEventOne.BidId = strconv.Itoa(bidId)
			flipKickErr := flipKickRepo.Create(headerOneId, []interface{}{flipKickEventOne})
			Expect(flipKickErr).NotTo(HaveOccurred())

			flipKickEventTwo := test_data.FlipKickModel
			flipKickEventTwo.ContractAddress = anotherFlipContractAddress
			flipKickEventTwo.BidId = strconv.Itoa(bidId)
			flipKickEventTwo.Lot = strconv.Itoa(differentLot)
			flipKickEventTwo.Bid = strconv.Itoa(differentBidAmount)
			flipKickEventTwo.TransactionIndex = flipKickEventOne.TransactionIndex + 1
			flipKickEventTwo.LogIndex = flipKickEventOne.LogIndex + 1
			flipKickErr = flipKickRepo.Create(headerOneId, []interface{}{flipKickEventTwo})
			Expect(flipKickErr).NotTo(HaveOccurred())

			var actualBidEvents []test_helpers.BidEvent
			queryErr := db.Select(&actualBidEvents, `SELECT bid_id, bid_amount, lot, act FROM api.all_flip_bid_events()`)
			Expect(queryErr).NotTo(HaveOccurred())

			Expect(actualBidEvents).To(ConsistOf(
				test_helpers.BidEvent{BidId: flipKickEventOne.BidId, BidAmount: flipKickEventOne.Bid, Lot: flipKickEventOne.Lot, Act: "kick"},
				test_helpers.BidEvent{BidId: flipKickEventTwo.BidId, BidAmount: flipKickEventTwo.Bid, Lot: flipKickEventTwo.Lot, Act: "kick"},
			))
		})

		Describe("tend", func() {
			It("returns tend events from multiple blocks", func() {
				lotOne := rand.Int()
				lotTwo := rand.Int()
				bidAmountOne := rand.Int()
				bidAmountTwo := rand.Int()

				headerOne := fakes.GetFakeHeader(1)
				headerOneId, headerOneErr := headerRepo.CreateOrUpdateHeader(headerOne)
				Expect(headerOneErr).NotTo(HaveOccurred())

				flipKickEvent := test_data.FlipKickModel
				flipKickEvent.ContractAddress = contractAddress
				flipKickEvent.BidId = strconv.Itoa(bidId)
				flipKickErr := flipKickRepo.Create(headerOneId, []interface{}{flipKickEvent})
				Expect(flipKickErr).NotTo(HaveOccurred())

				flipTendErr := test_helpers.CreateTend(test_helpers.TendCreationInput{
					BidId:           bidId,
					ContractAddress: contractAddress,
					Lot:             lotOne,
					BidAmount:       bidAmountOne,
					TendRepo:        tendRepo,
					TendHeaderId:    headerOneId,
				})
				Expect(flipTendErr).NotTo(HaveOccurred())

				headerTwo := fakes.GetFakeHeader(2)
				headerTwoId, headerTwoErr := headerRepo.CreateOrUpdateHeader(headerTwo)
				Expect(headerTwoErr).NotTo(HaveOccurred())

				flipTendErr = test_helpers.CreateTend(test_helpers.TendCreationInput{
					BidId:           bidId,
					ContractAddress: contractAddress,
					Lot:             lotTwo,
					BidAmount:       bidAmountTwo,
					TendRepo:        tendRepo,
					TendHeaderId:    headerTwoId,
				})
				Expect(flipTendErr).NotTo(HaveOccurred())

				var actualBidEvents []test_helpers.BidEvent
				queryErr := db.Select(&actualBidEvents, `SELECT bid_id, bid_amount, lot, act FROM api.all_flip_bid_events()`)
				Expect(queryErr).NotTo(HaveOccurred())

				Expect(actualBidEvents).To(ConsistOf(
					test_helpers.BidEvent{BidId: flipKickEvent.BidId, BidAmount: flipKickEvent.Bid, Lot: flipKickEvent.Lot, Act: "kick"},
					test_helpers.BidEvent{BidId: strconv.Itoa(bidId), BidAmount: strconv.Itoa(bidAmountOne), Lot: strconv.Itoa(lotOne), Act: "tend"},
					test_helpers.BidEvent{BidId: strconv.Itoa(bidId), BidAmount: strconv.Itoa(bidAmountTwo), Lot: strconv.Itoa(lotTwo), Act: "tend"},
				))
			})

			It("ignores tend events that are not from a flip", func() {
				headerOne := fakes.GetFakeHeader(1)
				headerOneId, headerOneErr := headerRepo.CreateOrUpdateHeader(headerOne)
				Expect(headerOneErr).NotTo(HaveOccurred())

				flipKickEvent := test_data.FlipKickModel
				flipKickEvent.ContractAddress = contractAddress
				flipKickEvent.BidId = strconv.Itoa(bidId)
				flipKickErr := flipKickRepo.Create(headerOneId, []interface{}{flipKickEvent})
				Expect(flipKickErr).NotTo(HaveOccurred())

				flapTendErr := test_helpers.CreateTend(test_helpers.TendCreationInput{
					BidId:           bidId,
					ContractAddress: "flap",
					Lot:             rand.Int(),
					BidAmount:       rand.Int(),
					TendRepo:        tendRepo,
					TendHeaderId:    headerOneId,
				})
				Expect(flapTendErr).NotTo(HaveOccurred())

				var actualBidEvents []test_helpers.BidEvent
				queryErr := db.Select(&actualBidEvents, `SELECT bid_id, bid_amount, lot, act FROM api.all_flip_bid_events()`)
				Expect(queryErr).NotTo(HaveOccurred())

				Expect(actualBidEvents).To(ConsistOf(
					test_helpers.BidEvent{BidId: flipKickEvent.BidId, BidAmount: flipKickEvent.Bid, Lot: flipKickEvent.Lot, Act: "kick"},
				))
			})
		})

		Describe("dent", func() {
			It("returns dent events from multiple blocks", func() {
				lotOne := rand.Int()
				lotTwo := rand.Int()
				bidAmountOne := rand.Int()
				bidAmountTwo := rand.Int()

				headerOne := fakes.GetFakeHeader(1)
				headerOneId, headerOneErr := headerRepo.CreateOrUpdateHeader(headerOne)
				Expect(headerOneErr).NotTo(HaveOccurred())

				flipKickEvent := test_data.FlipKickModel
				flipKickEvent.ContractAddress = contractAddress
				flipKickEvent.BidId = strconv.Itoa(bidId)
				flipKickErr := flipKickRepo.Create(headerOneId, []interface{}{flipKickEvent})
				Expect(flipKickErr).NotTo(HaveOccurred())

				flipDentErr := test_helpers.CreateDent(test_helpers.DentCreationInput{
					BidId:           bidId,
					ContractAddress: contractAddress,
					Lot:             lotOne,
					BidAmount:       bidAmountOne,
					DentRepo:        dentRepo,
					DentHeaderId:    headerOneId,
				})
				Expect(flipDentErr).NotTo(HaveOccurred())

				headerTwo := fakes.GetFakeHeader(2)
				headerTwoId, headerTwoErr := headerRepo.CreateOrUpdateHeader(headerTwo)
				Expect(headerTwoErr).NotTo(HaveOccurred())

				flipDentErr = test_helpers.CreateDent(test_helpers.DentCreationInput{
					BidId:           bidId,
					ContractAddress: contractAddress,
					Lot:             lotTwo,
					BidAmount:       bidAmountTwo,
					DentRepo:        dentRepo,
					DentHeaderId:    headerTwoId,
				})
				Expect(flipDentErr).NotTo(HaveOccurred())

				var actualBidEvents []test_helpers.BidEvent
				queryErr := db.Select(&actualBidEvents, `SELECT bid_id, bid_amount, lot, act FROM api.all_flip_bid_events()`)
				Expect(queryErr).NotTo(HaveOccurred())

				Expect(actualBidEvents).To(ConsistOf(
					test_helpers.BidEvent{BidId: flipKickEvent.BidId, BidAmount: flipKickEvent.Bid, Lot: flipKickEvent.Lot, Act: "kick"},
					test_helpers.BidEvent{BidId: strconv.Itoa(bidId), BidAmount: strconv.Itoa(bidAmountOne), Lot: strconv.Itoa(lotOne), Act: "dent"},
					test_helpers.BidEvent{BidId: strconv.Itoa(bidId), BidAmount: strconv.Itoa(bidAmountTwo), Lot: strconv.Itoa(lotTwo), Act: "dent"},
				))
			})

			It("ignores dent events that are not from flip", func() {
				headerOne := fakes.GetFakeHeader(1)
				headerOneId, headerOneErr := headerRepo.CreateOrUpdateHeader(headerOne)
				Expect(headerOneErr).NotTo(HaveOccurred())

				flipKickEvent := test_data.FlipKickModel
				flipKickEvent.ContractAddress = contractAddress
				flipKickEvent.BidId = strconv.Itoa(bidId)
				flipKickErr := flipKickRepo.Create(headerOneId, []interface{}{flipKickEvent})
				Expect(flipKickErr).NotTo(HaveOccurred())

				flapDentErr := test_helpers.CreateDent(test_helpers.DentCreationInput{
					BidId:           bidId,
					ContractAddress: "flap",
					Lot:             rand.Int(),
					BidAmount:       rand.Int(),
					DentRepo:        dentRepo,
					DentHeaderId:    headerOneId,
				})
				Expect(flapDentErr).NotTo(HaveOccurred())

				var actualBidEvents []test_helpers.BidEvent
				queryErr := db.Select(&actualBidEvents, `SELECT bid_id, bid_amount, lot, act FROM api.all_flip_bid_events()`)
				Expect(queryErr).NotTo(HaveOccurred())

				Expect(actualBidEvents).To(ConsistOf(
					test_helpers.BidEvent{BidId: flipKickEvent.BidId, BidAmount: flipKickEvent.Bid, Lot: flipKickEvent.Lot, Act: "kick"},
				))
			})
		})

		Describe("yank", func() {
			It("includes yank and gets values from the block where the yank occurred", func() {
				tendLot := rand.Int()
				tendBidAmount := rand.Int()

				headerOne := fakes.GetFakeHeader(1)
				headerOneId, headerOneErr := headerRepo.CreateOrUpdateHeader(headerOne)
				Expect(headerOneErr).NotTo(HaveOccurred())

				flipKickEvent := test_data.FlipKickModel
				flipKickEvent.ContractAddress = contractAddress
				flipKickEvent.BidId = strconv.Itoa(bidId)
				flipKickErr := flipKickRepo.Create(headerOneId, []interface{}{flipKickEvent})
				Expect(flipKickErr).NotTo(HaveOccurred())

				flipTendErr := test_helpers.CreateTend(test_helpers.TendCreationInput{
					BidId:           bidId,
					ContractAddress: contractAddress,
					Lot:             tendLot,
					BidAmount:       tendBidAmount,
					TendRepo:        tendRepo,
					TendHeaderId:    headerOneId,
				})
				Expect(flipTendErr).NotTo(HaveOccurred())

				flipStorageValues := test_helpers.GetFlipStorageValues(1, test_helpers.FakeIlk.Hex, bidId)
				test_helpers.CreateFlip(db, headerOne, flipStorageValues,
					test_helpers.GetFlipMetadatas(strconv.Itoa(bidId)), contractAddress)

				headerTwo := fakes.GetFakeHeader(2)
				headerTwoId, headerTwoErr := headerRepo.CreateOrUpdateHeader(headerTwo)
				Expect(headerTwoErr).NotTo(HaveOccurred())

				flipYankErr := test_helpers.CreateYank(test_helpers.YankCreationInput{
					BidId:           bidId,
					ContractAddress: contractAddress,
					YankRepo:        yankRepo,
					YankHeaderId:    headerTwoId,
				})
				Expect(flipYankErr).NotTo(HaveOccurred())

				updatedFlipStorageValues := test_helpers.GetFlipStorageValues(2, test_helpers.FakeIlk.Hex, bidId)
				test_helpers.CreateFlip(db, headerTwo, updatedFlipStorageValues,
					test_helpers.GetFlipMetadatas(strconv.Itoa(bidId)), contractAddress)

				var actualBidEvents []test_helpers.BidEvent
				queryErr := db.Select(&actualBidEvents, `SELECT bid_id, bid_amount, lot, act FROM api.all_flip_bid_events()`)
				Expect(queryErr).NotTo(HaveOccurred())

				Expect(actualBidEvents).To(ConsistOf(
					test_helpers.BidEvent{
						BidId:     flipKickEvent.BidId,
						BidAmount: flipKickEvent.Bid,
						Lot:       flipKickEvent.Lot,
						Act:       "kick",
					},
					test_helpers.BidEvent{
						BidId:     strconv.Itoa(bidId),
						BidAmount: strconv.Itoa(tendBidAmount),
						Lot:       strconv.Itoa(tendLot),
						Act:       "tend",
					},
					test_helpers.BidEvent{
						BidId:     strconv.Itoa(bidId),
						BidAmount: updatedFlipStorageValues[storage.BidBid].(string),
						Lot:       updatedFlipStorageValues[storage.BidLot].(string),
						Act:       "yank",
					},
				))
			})
		})
	})
})